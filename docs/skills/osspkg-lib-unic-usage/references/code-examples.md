# go-unic code references

The examples below are intentionally small patterns for code review and
implementation. They reflect the current public API.

## Correct: UNIC configuration shape

This configuration matches the typed example above. Repeated block names map
to a slice, values before `{` map to fields tagged with `attr=N`, and maps use
an even sequence of key/value items.

```unic
log_level info;
server api 8080 {
	tags [public, http];
	env (MODE, production, REGION, eu-west);
}
server admin 9090 {
	tags [internal];
	env (MODE, staging);
}
```

Strings containing spaces or syntax characters must be quoted. Triple
backticks are used when both quote styles or line breaks are needed.

```unic
title 'Production service';
path '/srv/app;current';
message ```line one
line two```;
```

Comments start with `#` and continue to the end of the line:

```unic
server api 8080 { # public HTTP server
	port 8080; # override the default
}
```

## Incorrect: malformed or mismatched configuration

```unic
# Wrong: a block field needs a closing brace.
server api 8080 {
	port 8080;

# Wrong: a list requires commas and a terminating semicolon.
tags [public internal]

# Wrong: map items must be key/value pairs.
env (MODE, production, REGION);
```

Other common mistakes are using an attribute without a matching `attr=N` tag,
putting a scalar where the Go field is a struct, and using a field name that is
not present in the target type when the value is expected to be decoded. Unknown
fields are ignored by the decoder, so application-required fields must be
validated after `Unmarshal`.

```unic
# Wrong for `Name string `unic:"name,attr=1"```: the first attribute is missing.
server { port 8080; }

# Wrong for `Tags []string`: this is a scalar, not a list.
tags public;
```

## Correct: typed configuration loading

```go
type Server struct {
	Name string            `unic:"name,attr=1"`
	Port int               `unic:"port,default=8080"`
	Tags []string          `unic:"tags,omitempty"`
	Env  map[string]string `unic:"env,omitempty"`
}

type Config struct {
	LogLevel string   `unic:"log_level,default='info'"`
	Servers  []Server `unic:"server"`
}

func loadConfig(data []byte) (Config, error) {
	var cfg Config
	if err := unic.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("load config: %w", err)
	}
	if len(cfg.Servers) == 0 {
		return Config{}, errors.New("config must define at least one server")
	}
	return cfg, nil
}
```

## Correct: serialization with error handling

```go
data, err := unic.Marshal(&cfg)
if err != nil {
	return fmt.Errorf("serialize config: %w", err)
}
if err := os.WriteFile("config.unic", data, 0o600); err != nil {
	return fmt.Errorf("write config: %w", err)
}
```

## Correct: dynamic data only where needed

```go
type Config struct {
	Metadata map[string]any `unic:"metadata,omitempty"`
}

var cfg Config
if err := unic.Unmarshal(data, &cfg); err != nil {
	return err
}
```

## Incorrect: passing a struct value to `Unmarshal`

```go
var cfg Config
_ = unic.Unmarshal(data, cfg) // wrong: Unmarshal requires *struct
```

The decoder must be able to update the destination. Use `&cfg` and do not pass
`nil` or a non-struct pointer.

## Incorrect: assuming a top-level map can be marshaled

```go
data, err := unic.Marshal(map[string]string{"host": "localhost"}) // wrong
```

Wrap the map in a tagged struct if it is the document's field:

```go
type Document struct {
	Values map[string]string `unic:"values"`
}

data, err := unic.Marshal(Document{
	Values: map[string]string{"host": "localhost"},
})
```

## Incorrect: ignoring conversion and I/O errors

```go
data, _ := unic.Marshal(cfg)
_ = os.WriteFile("config.unic", data, 0o600)
```

Ignoring errors can turn an unsupported field type, invalid tag, or failed
write into a silently incomplete configuration. Return or handle each error.

## Incorrect: treating tags as application validation

```go
type Config struct {
	Port int `unic:"port,default=8080"`
}
```

The default supplies a missing value, but does not enforce a valid port range.
Validate conditions such as `1 <= Port <= 65535` after unmarshalling.
