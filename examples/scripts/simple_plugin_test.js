describe('Simple Plugin Test', () => {
  test('math plugin basic test', () => {
    const math = require('./plugins/examples/math/math.so');
    expect(math.add(1, 2)).toBe(3);
    expect(math.add(1, 3)).toBe(4);
  });
});