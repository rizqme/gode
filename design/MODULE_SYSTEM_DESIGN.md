# ES6 Module System Design for Goja + Gode

## Overview

This document outlines the design for adding ES6 import/export syntax support to the goja JavaScript engine, with module resolution delegated to the Gode runtime.

## Architecture

### 1. Component Separation

- **Goja**: Handles import/export syntax parsing and AST generation
- **Gode**: Handles module resolution, loading, and caching
- **Bridge**: Interface between goja and Gode for module operations

### 2. Module Resolution Interface

```go
// ModuleResolver interface that Gode implements
type ModuleResolver interface {
    // Resolve a module specifier to a resolved path
    ResolveModule(specifier string, referrer string) (string, error)
    
    // Load module source code
    LoadModule(path string) (string, error)
    
    // Get module exports for completed modules
    GetModuleExports(path string) (interface{}, error)
}

// ImportBinding represents an import specifier
type ImportBinding struct {
    Local    string // Local name
    Imported string // Imported name ("default" for default imports)
    Source   string // Module specifier
}

// ExportBinding represents an export specifier
type ExportBinding struct {
    Local    string // Local name
    Exported string // Exported name ("default" for default exports)
    Source   string // Module specifier (for re-exports)
}
```

## Implementation Plan

### Phase 1: Token and AST Extensions

#### 1.1 Token Updates (`goja/token/token_const.go`)

Add dedicated IMPORT and EXPORT tokens:

```go
// Add to token constants (line ~127)
IMPORT      // import
EXPORT      // export

// Add to string mappings (line ~241)  
IMPORT:      "import",
EXPORT:      "export",

// Update keyword table (lines 346-349)
"export": {
    token: EXPORT,  // Change from KEYWORD
},
"import": {
    token: IMPORT,  // Change from KEYWORD
},
```

#### 1.2 AST Node Extensions (`goja/ast/node.go`)

Add import/export AST nodes:

```go
// Import statement types
ImportDeclaration struct {
    Import     file.Idx
    Specifiers []ImportSpecifier
    Source     *StringLiteral
}

ImportSpecifier struct {
    Local    *Identifier
    Imported *Identifier  // nil for namespace imports
    IsDefault bool
}

// Export statement types
ExportDeclaration struct {
    Export      file.Idx
    Declaration Statement    // For export declarations
    Specifiers  []ExportSpecifier
    Source      *StringLiteral // For re-exports
    IsDefault   bool
}

ExportSpecifier struct {
    Local    *Identifier
    Exported *Identifier
}

// Dynamic import expression
ImportExpression struct {
    Import file.Idx
    Source Expression
}
```

### Phase 2: Parser Extensions

#### 2.1 Statement Parsing (`goja/parser/statement.go`)

Add import/export parsing to `parseStatement()`:

```go
case token.IMPORT:
    return self.parseImportStatement()
case token.EXPORT:
    return self.parseExportStatement()
```

#### 2.2 Import Parser Implementation

```go
func (self *_parser) parseImportStatement() ast.Statement {
    idx := self.expect(token.IMPORT)
    
    // Handle dynamic imports: import(specifier)
    if self.token == token.LEFT_PARENTHESIS {
        return self.parseImportExpression(idx)
    }
    
    // Static imports
    var specifiers []ast.ImportSpecifier
    var source *ast.StringLiteral
    
    if self.token == token.STRING {
        // import "module" - side effect only
        source = &ast.StringLiteral{
            Idx:     self.idx,
            Literal: self.literal,
            Value:   self.value,
        }
        self.next()
    } else {
        // Parse import specifiers
        specifiers = self.parseImportSpecifiers()
        self.expect(token.IDENTIFIER) // "from"
        source = self.parseStringLiteral()
    }
    
    self.semicolon()
    
    return &ast.ImportDeclaration{
        Import:     idx,
        Specifiers: specifiers,
        Source:     source,
    }
}
```

#### 2.3 Export Parser Implementation

```go
func (self *_parser) parseExportStatement() ast.Statement {
    idx := self.expect(token.EXPORT)
    
    // export default ...
    if self.token == token.DEFAULT {
        self.next()
        return self.parseExportDefault(idx)
    }
    
    // export { ... } from "module"
    if self.token == token.LEFT_BRACE {
        return self.parseExportSpecifiers(idx)
    }
    
    // export declaration
    return self.parseExportDeclaration(idx)
}
```

### Phase 3: Compiler Extensions

#### 3.1 Module Context (`goja/compiler_module.go` - new file)

```go
type ModuleCompiler struct {
    *compiler
    resolver ModuleResolver
    imports  []ImportBinding
    exports  []ExportBinding
}

func (c *ModuleCompiler) compileImportDeclaration(v *ast.ImportDeclaration) {
    // Resolve module path
    resolved, err := c.resolver.ResolveModule(v.Source.Value, c.file.name)
    if err != nil {
        c.error(v.Source.Idx0(), err.Error())
        return
    }
    
    // Register import bindings
    for _, spec := range v.Specifiers {
        binding := ImportBinding{
            Local:    spec.Local.Name,
            Imported: spec.Imported.Name,
            Source:   resolved,
        }
        c.imports = append(c.imports, binding)
    }
    
    // Generate module loading code
    c.compileModuleLoad(resolved, v.Specifiers)
}

func (c *ModuleCompiler) compileExportDeclaration(v *ast.ExportDeclaration) {
    // Handle different export types
    if v.Declaration != nil {
        // export function foo() {} / export const x = 1
        c.compileStatement(v.Declaration, false)
        c.registerExportFromDeclaration(v.Declaration)
    } else {
        // export { ... } from "module"
        c.compileExportSpecifiers(v.Specifiers, v.Source)
    }
}
```

#### 3.2 Runtime Module Loading

```go
func (c *ModuleCompiler) compileModuleLoad(resolved string, specifiers []ast.ImportSpecifier) {
    // Generate code equivalent to:
    // const __module = require(resolved);
    // const localName = __module.exportedName;
    
    c.emit(loadModule, resolved) // Custom opcode
    
    for _, spec := range specifiers {
        if spec.IsDefault {
            c.emit(loadDefault, spec.Local.Name)
        } else {
            c.emit(loadNamed, spec.Local.Name, spec.Imported.Name)
        }
    }
}
```

### Phase 4: VM Extensions

#### 4.1 Module Instructions (`goja/vm.go`)

Add new opcodes for module operations:

```go
const (
    // ... existing opcodes
    loadModule   _opcode = iota
    loadDefault
    loadNamed
    exportNamed
    exportDefault
)

func (vm *vm) runLoadModule() {
    specifier := vm.r(vm.pc).str()
    vm.pc++
    
    // Call module resolver
    if resolver, ok := vm.r.moduleResolver.(ModuleResolver); ok {
        exports, err := resolver.GetModuleExports(specifier)
        if err != nil {
            vm.throw(vm.r.NewGoError(err))
            return
        }
        vm.push(vm.r.ToValue(exports))
    } else {
        vm.throw(vm.r.NewGoError(fmt.Errorf("module resolver not available")))
    }
}
```

### Phase 5: Gode Integration

#### 5.1 Runtime Module Resolver (`internal/runtime/module_resolver.go`)

```go
type ModuleResolver struct {
    runtime *Runtime
    manager *modules.ModuleManager
}

func (r *ModuleResolver) ResolveModule(specifier string, referrer string) (string, error) {
    return r.manager.Resolve(specifier, referrer)
}

func (r *ModuleResolver) LoadModule(path string) (string, error) {
    return r.manager.Load(path)
}

func (r *ModuleResolver) GetModuleExports(path string) (interface{}, error) {
    // Load module if not already loaded
    source, err := r.LoadModule(path)
    if err != nil {
        return nil, err
    }
    
    if source == "" {
        // Plugin or built-in module
        return r.runtime.modules[path], nil
    }
    
    // Execute module and return exports
    return r.executeModule(path, source)
}
```

#### 5.2 Module Execution Context

```go
func (r *ModuleResolver) executeModule(path string, source string) (interface{}, error) {
    // Create module scope
    moduleScope := fmt.Sprintf(`
        (function(exports, require, module, __filename, __dirname) {
            %s
            return exports;
        })
    `, source)
    
    // Execute in module context
    done := make(chan interface{}, 1)
    r.runtime.QueueJSOperation(func() {
        exports := r.runtime.runtime.NewObject()
        module := r.runtime.runtime.NewObject()
        module.Set("exports", exports)
        
        fn, err := r.runtime.runtime.RunString(moduleScope)
        if err != nil {
            done <- err
            return
        }
        
        result, err := fn.ToObject(r.runtime.runtime).Call(goja.Undefined(), 
            exports, r.runtime.runtime.Get("require"), module, path, filepath.Dir(path))
        if err != nil {
            done <- err
            return
        }
        
        done <- result
    })
    
    result := <-done
    if err, ok := result.(error); ok {
        return nil, err
    }
    
    return result, nil
}
```

## Implementation Priority

1. **Phase 1**: Token and AST extensions (Foundation)
2. **Phase 2**: Parser extensions (Syntax support)
3. **Phase 3**: Compiler extensions (Code generation)
4. **Phase 4**: VM extensions (Runtime execution)
5. **Phase 5**: Gode integration (Module resolution)

## Testing Strategy

### Unit Tests
- AST node serialization/deserialization
- Parser import/export syntax validation
- Module resolution edge cases

### Integration Tests
- Import/export between JavaScript modules
- Mixed CommonJS/ES6 module usage
- Plugin imports via ES6 syntax
- Circular dependency handling

### Example Test Cases

```javascript
// Basic named imports
import { add, subtract } from './math.js';

// Default imports
import Calculator from './calculator.js';

// Namespace imports
import * as utils from './utils.js';

// Re-exports
export { add as sum } from './math.js';
export * from './helpers.js';

// Dynamic imports
const math = await import('./math.js');
```

## Future Enhancements

1. **Top-level await** - Module-level async execution
2. **Import maps** - Advanced module specifier resolution
3. **Module federation** - Cross-runtime module sharing
4. **Tree shaking** - Dead code elimination in builds
5. **Hot module replacement** - Development-time module updates

## Compatibility

- **Maintains backward compatibility** with existing CommonJS `require()` system
- **Interoperability** between ES6 modules and CommonJS modules
- **Progressive enhancement** - ES6 syntax availability without breaking existing code
- **Standard compliance** - Follows ES6 module specification where possible