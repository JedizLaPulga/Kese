# Kese

> A modern, fast, and effective Go web framework inspired by FastAPI

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://go.dev/)

## 🎯 Overview

Kese is a lightweight, high-performance web framework for Go that brings the elegant developer experience of FastAPI to the Go ecosystem. Built entirely on Go's standard library, Kese provides powerful features without getting in your way.

### Philosophy

**Do much, but stay out of the way.**

Kese is designed to give you all the tools you need to build modern web applications while maintaining the simplicity and performance that makes Go great.

## ✨ Features

- 🚀 **Fast**: Built on Go's standard library for maximum performance
- 🎨 **Elegant**: Clean, intuitive API inspired by FastAPI
- 📦 **Zero Dependencies**: Uses only the Go standard library
- 🧪 **Fully Tested**: Comprehensive test coverage
- 🔧 **Modular**: Pick and choose the components you need
- 📝 **Type-Safe**: Leverage Go's type system for safer code
- 🎯 **Developer-Friendly**: Minimal boilerplate, maximum productivity

## 🚧 Project Status

**Kese is currently under active development.** We're building a solid foundation before the first release.

## 🏗️ Core Components (Planned)

- **Router**: Fast, flexible HTTP routing
- **Request/Response Handling**: Intuitive request parsing and response generation
- **Middleware**: Composable middleware system
- **Validation**: Built-in request validation
- **Documentation**: Auto-generated API documentation
- **Dependency Injection**: Simple and effective DI system

## 🎓 Quick Start (Coming Soon)

```go
package main

import (
    "github.com/JedizLaPulga/kese"
)

func main() {
    app := kese.New()
    
    app.GET("/", func(c *kese.Context) error {
        return c.JSON(200, map[string]string{
            "message": "Hello from Kese!",
        })
    })
    
    app.Run(":8080")
}
```

## 🗺️ Roadmap

- [ ] Core routing engine
- [ ] Request/response handling
- [ ] Middleware system
- [ ] Context management
- [ ] Request validation
- [ ] Error handling
- [ ] Testing utilities
- [ ] Documentation generation
- [ ] Real-world validation project

## 🤝 Contributing

Contributions are welcome! Please read our contributing guidelines before submitting PRs.

### Development Principles

1. **Standard Practices**: Follow Go idioms and best practices
2. **Test Everything**: All code must include tests
3. **Stay Modular**: Keep components focused and decoupled
4. **Document Well**: Code and features should be well-documented

See [PROJECT_RULES.md](PROJECT_RULES.md) for detailed development guidelines.

## 📋 Requirements

- Go 1.21 or higher
- No external dependencies required

## 📖 Documentation

Documentation will be available once the core framework is complete.

## 🧪 Testing

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with verbose output
go test -v ./...
```

## 📜 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

Inspired by [FastAPI](https://fastapi.tiangolo.com/) - bringing Python's most elegant web framework patterns to Go.

## 💬 Community

- **Issues**: Use GitHub Issues for bug reports and feature requests
- **Discussions**: Join our GitHub Discussions for questions and ideas

---

**Built with ❤️ and Go**
