package globals

import (
	"fmt"
	"net/url"
	"strings"
)

// URL represents a URL object similar to the Web API
type URL struct {
	parsed       *url.URL
	searchParams *URLSearchParams
}

// URLConstructor provides the URL constructor function
type URLConstructor struct{}

// NewURL creates a new URL object
func (uc *URLConstructor) New(input string, base ...string) (*URL, error) {
	var parsed *url.URL
	var err error
	
	if len(base) > 0 && base[0] != "" {
		// Parse with base URL
		baseURL, err := url.Parse(base[0])
		if err != nil {
			return nil, fmt.Errorf("Invalid base URL")
		}
		parsed, err = baseURL.Parse(input)
		if err != nil {
			return nil, fmt.Errorf("Invalid URL")
		}
	} else {
		// Parse absolute URL
		parsed, err = url.Parse(input)
		if err != nil {
			return nil, fmt.Errorf("Invalid URL")
		}
		if !parsed.IsAbs() {
			return nil, fmt.Errorf("Invalid URL")
		}
	}
	
	u := &URL{
		parsed: parsed,
	}
	u.searchParams = NewURLSearchParams(parsed.RawQuery)
	u.searchParams.url = u
	
	return u, nil
}

// Properties

func (u *URL) Href() string {
	return u.parsed.String()
}

func (u *URL) SetHref(href string) error {
	parsed, err := url.Parse(href)
	if err != nil {
		return err
	}
	u.parsed = parsed
	u.searchParams = NewURLSearchParams(parsed.RawQuery)
	u.searchParams.url = u
	return nil
}

func (u *URL) Origin() string {
	if u.parsed.Scheme == "" {
		return ""
	}
	return fmt.Sprintf("%s://%s", u.parsed.Scheme, u.parsed.Host)
}

func (u *URL) Protocol() string {
	return u.parsed.Scheme + ":"
}

func (u *URL) SetProtocol(protocol string) {
	u.parsed.Scheme = strings.TrimSuffix(protocol, ":")
}

func (u *URL) Username() string {
	if u.parsed.User != nil {
		return u.parsed.User.Username()
	}
	return ""
}

func (u *URL) SetUsername(username string) {
	password, _ := u.parsed.User.Password()
	if username == "" && password == "" {
		u.parsed.User = nil
	} else {
		u.parsed.User = url.UserPassword(username, password)
	}
}

func (u *URL) Password() string {
	if u.parsed.User != nil {
		password, _ := u.parsed.User.Password()
		return password
	}
	return ""
}

func (u *URL) SetPassword(password string) {
	username := ""
	if u.parsed.User != nil {
		username = u.parsed.User.Username()
	}
	if username == "" && password == "" {
		u.parsed.User = nil
	} else {
		u.parsed.User = url.UserPassword(username, password)
	}
}

func (u *URL) Host() string {
	return u.parsed.Host
}

func (u *URL) SetHost(host string) {
	u.parsed.Host = host
}

func (u *URL) Hostname() string {
	return u.parsed.Hostname()
}

func (u *URL) SetHostname(hostname string) {
	port := u.parsed.Port()
	if port != "" {
		u.parsed.Host = hostname + ":" + port
	} else {
		u.parsed.Host = hostname
	}
}

func (u *URL) Port() string {
	return u.parsed.Port()
}

func (u *URL) SetPort(port string) {
	hostname := u.parsed.Hostname()
	if port != "" {
		u.parsed.Host = hostname + ":" + port
	} else {
		u.parsed.Host = hostname
	}
}

func (u *URL) Pathname() string {
	if u.parsed.Path == "" {
		return "/"
	}
	return u.parsed.Path
}

func (u *URL) SetPathname(pathname string) {
	u.parsed.Path = pathname
}

func (u *URL) Search() string {
	if u.parsed.RawQuery == "" {
		return ""
	}
	return "?" + u.parsed.RawQuery
}

func (u *URL) SetSearch(search string) {
	u.parsed.RawQuery = strings.TrimPrefix(search, "?")
	u.searchParams = NewURLSearchParams(u.parsed.RawQuery)
	u.searchParams.url = u
}

func (u *URL) SearchParams() *URLSearchParams {
	return u.searchParams
}

func (u *URL) Hash() string {
	if u.parsed.Fragment == "" {
		return ""
	}
	return "#" + u.parsed.Fragment
}

func (u *URL) SetHash(hash string) {
	u.parsed.Fragment = strings.TrimPrefix(hash, "#")
}

// Methods

func (u *URL) ToString() string {
	return u.Href()
}

func (u *URL) ToJSON() string {
	return u.Href()
}

// URLSearchParams represents URL query parameters
type URLSearchParams struct {
	params [][]string // Array of [key, value] pairs to maintain order
	url    *URL       // Reference to parent URL if any
}

// NewURLSearchParams creates a new URLSearchParams object
func NewURLSearchParams(init ...string) *URLSearchParams {
	usp := &URLSearchParams{
		params: make([][]string, 0),
	}
	
	if len(init) > 0 && init[0] != "" {
		// Parse query string
		values, _ := url.ParseQuery(init[0])
		for key, vals := range values {
			for _, val := range vals {
				usp.params = append(usp.params, []string{key, val})
			}
		}
	}
	
	return usp
}

// Append adds a new value to the params
func (usp *URLSearchParams) Append(name, value string) {
	usp.params = append(usp.params, []string{name, value})
	usp.updateURL()
}

// Delete removes all values for a key
func (usp *URLSearchParams) Delete(name string) {
	newParams := make([][]string, 0)
	for _, param := range usp.params {
		if param[0] != name {
			newParams = append(newParams, param)
		}
	}
	usp.params = newParams
	usp.updateURL()
}

// Get returns the first value for a key
func (usp *URLSearchParams) Get(name string) string {
	for _, param := range usp.params {
		if param[0] == name {
			return param[1]
		}
	}
	return ""
}

// GetAll returns all values for a key
func (usp *URLSearchParams) GetAll(name string) []string {
	values := make([]string, 0)
	for _, param := range usp.params {
		if param[0] == name {
			values = append(values, param[1])
		}
	}
	return values
}

// Has checks if a key exists
func (usp *URLSearchParams) Has(name string) bool {
	for _, param := range usp.params {
		if param[0] == name {
			return true
		}
	}
	return false
}

// Set sets the value for a key (removes all existing values first)
func (usp *URLSearchParams) Set(name, value string) {
	found := false
	newParams := make([][]string, 0)
	
	for _, param := range usp.params {
		if param[0] == name {
			if !found {
				newParams = append(newParams, []string{name, value})
				found = true
			}
		} else {
			newParams = append(newParams, param)
		}
	}
	
	if !found {
		newParams = append(newParams, []string{name, value})
	}
	
	usp.params = newParams
	usp.updateURL()
}

// Sort sorts all key/value pairs by key
func (usp *URLSearchParams) Sort() {
	// Bubble sort for simplicity (could be optimized)
	n := len(usp.params)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if usp.params[j][0] > usp.params[j+1][0] {
				usp.params[j], usp.params[j+1] = usp.params[j+1], usp.params[j]
			}
		}
	}
	usp.updateURL()
}

// ToString returns the query string
func (usp *URLSearchParams) ToString() string {
	if len(usp.params) == 0 {
		return ""
	}
	
	parts := make([]string, 0, len(usp.params))
	for _, param := range usp.params {
		parts = append(parts, url.QueryEscape(param[0])+"="+url.QueryEscape(param[1]))
	}
	return strings.Join(parts, "&")
}

// ForEach iterates over all key/value pairs
func (usp *URLSearchParams) ForEach(callback func(value, key string)) {
	for _, param := range usp.params {
		callback(param[1], param[0])
	}
}

// Keys returns an iterator of all keys
func (usp *URLSearchParams) Keys() []string {
	keys := make([]string, 0, len(usp.params))
	for _, param := range usp.params {
		keys = append(keys, param[0])
	}
	return keys
}

// Values returns an iterator of all values
func (usp *URLSearchParams) Values() []string {
	values := make([]string, 0, len(usp.params))
	for _, param := range usp.params {
		values = append(values, param[1])
	}
	return values
}

// Entries returns an iterator of all key/value pairs
func (usp *URLSearchParams) Entries() [][]string {
	entries := make([][]string, len(usp.params))
	copy(entries, usp.params)
	return entries
}

// updateURL updates the parent URL if attached
func (usp *URLSearchParams) updateURL() {
	if usp.url != nil {
		usp.url.parsed.RawQuery = usp.ToString()
	}
}