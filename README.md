# Universal Configuration Format (UNIC)

![GitHub Release](https://img.shields.io/github/v/release/osspkg/go-unic)
![License](https://img.shields.io/badge/license-BSD--3--Clause-blue)
[![GoDoc](https://pkg.go.dev/badge/go.osspkg.com/unic)](https://pkg.go.dev/go.osspkg.com/unic)
[![RU Lang](https://img.shields.io/badge/lang-RU-green?style=flat)](README.ru.md)


**go-unic** is a reliable, high‑performance Go library for parsing, serialising, and working with configuration files in the **Universal Configuration Format (UNIC)**. UNIC combines readability, a minimalistic syntax, and flexibility – letting you describe complex data structures as easily as with JSON or YAML, but with a syntax that feels more natural for humans.

## 📋 Table of Contents

- [Features](#-features)
- [Installation](#-installation)
- [Quick Start](#-quick-start)
    - [Unmarshal (deserialisation)](#1-unmarshal-deserialisation)
    - [Marshal (serialisation)](#2-marshal-serialisation)
- [UNIC Syntax](#-unic-syntax)
    - [Basic constructs](#basic-constructs)
    - [String escaping](#string-escaping)
    - [Comments](#comments)
    - [Attributes](#attributes)
- [Struct Tags (options)](#-struct-tags-options)
- [Examples](#-examples)
- [Comparison with other formats](#-comparison-with-other-formats)
- [Contributing](#-contributing)
- [License](#-license)

---

## 🚀 Features

- **Human‑readable syntax** – intuitive, without superfluous symbols (like Nginx or HCL).
- **Support for all major Go types**: structs, slices, maps, scalars (numbers, strings, booleans).
- **Flexible tag‑based control** – set field names, default values, omit empty fields, attributes, comments.
- **Structure merging** – automatically combine fields when serialising multiple objects with the same key.
- **Arbitrary nesting depth** – blocks, lists, and maps can be combined freely.
- **Support for `any` interfaces** – deserialise into `map[string]any` or `[]any` when needed.
- **High performance** – minimal reflection usage, cached struct metadata.
- **Zero allocations** during parsing (uses a `bb` buffer).

---

## 📦 Installation

```bash
go get -u go.osspkg.com/unic@latest
```

---

## 💡 Quick Start

### 1. Unmarshal (deserialisation)

To convert UNIC data into a Go struct, use `unic.Unmarshal`:

```go
package main

import (
	"fmt"
	"os"

	"go.osspkg.com/unic"
)

type Config struct {
	LogLevel int      `unic:"log_level,default=1"`
	Port     int      `unic:"port"`
	Features []string `unic:"features,omitempty"`
}

func main() {
	data, err := os.ReadFile("config.unic")
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		return
	}

	var cfg Config
	if err := unic.Unmarshal(data, &cfg); err != nil {
		fmt.Printf("Parsing error: %v\n", err)
		return
	}

	fmt.Printf("Configuration: log_level=%d, port=%d, features=%v\n",
		cfg.LogLevel, cfg.Port, cfg.Features)
}
```

### 2. Marshal (serialisation)

To write a struct to UNIC format, use `unic.Marshal`:

```go
package main

import (
	"fmt"
	"os"

	"go.osspkg.com/unic"
)

func main() {
	cfg := Config{
		LogLevel: 2,
		Port:     8080,
		Features: []string{"auth", "metrics"},
	}

	data, err := unic.Marshal(cfg)
	if err != nil {
		fmt.Printf("Serialisation error: %v\n", err)
		return
	}

	if err := os.WriteFile("config.unic", data, 0644); err != nil {
		fmt.Printf("Write error: %v\n", err)
	}
}
```

---

## 📐 UNIC Syntax

UNIC is a text file containing fields, blocks, lists, and maps. Basic rules:

### Basic constructs

| Construct      | Example                             | Description                                                    |
|----------------|-------------------------------------|----------------------------------------------------------------|
| **Field**      | `key value;`                        | Assigns a scalar value (string, number, boolean).              |
| **Block**      | `key { field1 val1; field2 val2; }` | Groups fields (similar to a struct).                           |
| **List**       | `key [val1, val2, val3];`           | Ordered collection of values.                                  |
| **Map**        | `key (key1, val1, key2, val2);`     | Key‑value pairs (keys are always strings).                     |
| **Attributes** | `key attr1 attr2 { ... }`           | Values before an opening block brace become struct attributes. |

### String escaping

To avoid conflicts with system characters (`{}[]();,#`), spaces, quotes, or line breaks, the following rules apply:

- If the string contains `"`, `{`, `}`, `[`, `]`, `(`, `)`, `#`, `;`, `,` or spaces – enclose it in single quotes: `'hello "world"'`.
- If the string contains `'`, `{`, `}`, `[`, `]`, `(`, `)`, `#`, `;`, `,` or spaces – enclose it in double quotes: `"hello 'world'"`.
- If the string contains both `'` and `"` as well as special characters or line breaks – use triple backticks: `` ```hello 'world' "foo"``` ``.

Example:

```
message 'Hello, "friend"!';
path "C:\\Program Files\\App";
multiline ```first line
second line```;
```

### Comments

- **Single‑line** – after `;` on the same line:
  ```
  port 80; # standard port
  ```
- **Block** – after `{` on the same line (applies to the entire block):
  ```
  server { # server settings
      host '127.0.0.1';
  }
  ```

### Attributes

If values are given before an opening block brace, they are interpreted as struct attributes. The `attr=N` tag sets the ordinal number (starting from 1).
Example:

```
server web 80 { host 'localhost'; }
```

corresponds to the struct:

```go
package main

type Server struct {
	Tag  string `unic:"tag,attr=1"`  // "web"
	Port int    `unic:"port,attr=2"` // 80
	Host string `unic:"host"`
}
```

---

## 🏷️ Struct Tags (options)

The `unic` tag has the format: `unic:"name[,option1='value'][,option2='value']..."`

Available options:

| Option      | Description                                                                                  |
|-------------|----------------------------------------------------------------------------------------------|
| `name`      | Field name in the configuration (mandatory).                                                 |
| `default`   | Default value (for scalars or lists separated by `;`).                                       |
| `omitempty` | If the field is empty (zero value), it is omitted during serialisation and ignored on parse. |
| `attr`      | Ordinal number of a block attribute (number > 0).                                            |
| `desc`      | Comment added during serialisation.                                                          |

Example:

```go
package main

type Config struct {
	LogLevel int      `unic:"log_level,default=1,desc='log level'"`
	Servers  []Server `unic:"server"`
}

type Server struct {
	Name string `unic:"name,attr=1"`
	Port int    `unic:"port"`
}
```

---

## 🔍 Examples

### Nested structs and lists

Configuration:

```
log_level 1;
servers {
    server web {
        port 80;
        host 'localhost';
    }
    server admin {
        port 8080;
        host '127.0.0.1';
        auth (user1, passwd1, user2, passwd2);
    }
}
```

Go struct:

```go
package main

type Config struct {
	LogLevel int `unic:"log_level"`
	Servers  struct {
		Servers []struct {
			Name string            `unic:"name,attr=1"`
			Port int               `unic:"port"`
			Host string            `unic:"host"`
			Auth map[string]string `unic:"auth,omitempty"`
		} `unic:"server"`
	} `unic:"servers"`
}
```

### Serialising multiple structs into one file

`unic.Marshal` accepts several arguments – all are merged into one document. If fields with the same name appear in different structs, they are combined (merged) into a single block.

```go
package main

type Part1 struct {
	Common string `unic:"common"`
	A      int    `unic:"a"`
}
type Part2 struct {
	Common string `unic:"common"`
	B      bool   `unic:"b"`
}
data, _ := unic.Marshal(Part1{Common:"shared", A:42}, Part2{Common:"shared", B:true})
// Output:
// common shared;
// a 42;
// b true;
```

---

## ⚖️ Comparison with other formats

| Format   | Readability | Complex structures | Comments | Performance (Go)              |
|----------|-------------|---------------------|----------|-------------------------------|
| **UNIC** | ★★★★★       | ★★★★★               | ✔        | High (reflection with cache)  |
| JSON     | ★★★☆☆       | ★★★★☆               | ✘        | Very high                     |
| YAML     | ★★★★☆       | ★★★★★               | ✔        | Medium                        |
| TOML     | ★★★★☆       | ★★★☆☆               | ✔        | Medium                        |
| HCL      | ★★★★☆       | ★★★★★               | ✔        | Medium                        |

**UNIC** offers the best balance between readability and performance, especially if you need comments, attributes, and flexible struct merging.

---

## 🤝 Contributing

We welcome your ideas and improvements! To contribute:

1. Fork the repository.
2. Create a branch for your feature (`git checkout -b feature/amazing-feature`).
3. Make your changes and write tests.
4. Ensure all linters and tests pass (`make pre-commit`).
5. Open a pull request.

---

## 📄 License

Distributed under the **BSD 3‑Clause** License. See the [LICENSE](LICENSE) file for details.
