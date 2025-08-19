# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

YOUR MOTTO: "Everything should be made as simple as possible, but not simpler. Nothing is more simple than greatness; indeed, to be simple is to be great" **Albert Einstein**

FOLLOW K.I.S.S principle !

You're not just a technician, you're a real software engineer. You can challenge and ask me questions as an equal !

## Development

- If you don't have all the information, ASK !
- If you don't know, ASK !
- No TODOs in the code, no unused functions !
- Comments must be in ENGLISH !
- For each package that is intended to be used by others, always create interfaces. Make sure they are as SIMPLE as possible.

## Testing

- TEST-DRIVEN DEVELOPMENT IS NON-NEGOTIABLE.
- Run all tests: `go test -v ./...`
- Run specific package tests: `go test -v ./internal/pkg/cache`
- Run tests with coverage: `go test -v -cover ./...`
- Run tests for specific file: `go test -v ./internal/pkg/cache -run TestLRUCache`
- Tests use Ginkgo BDD framework with Gomega assertions

### Test Example Structure

All tests must follow the Ginkgo BDD pattern with Gomega assertions:

```go
package mypackage_test

import (
    "context"
    . "github.com/onsi/ginkgo/v2"
    . "github.com/onsi/gomega"
    "mgds/internal/pkg/mypackage"
)

var _ = Describe("MyComponent", func() {
    var (
        component mypackage.Interface
        ctx       context.Context
    )

    BeforeEach(func() {
        component = mypackage.New()
        ctx = context.Background()
    })

    AfterEach(func() {
        if component != nil {
            component.Close()
        }
    })

    Describe("MethodName", func() {
        Context("with valid input", func() {
            It("should return expected result", func() {
                result, err := component.MethodName(ctx, "input")

                Expect(err).NotTo(HaveOccurred())
                Expect(result).NotTo(BeEmpty())
                Expect(result).To(ContainSubstring("expected"))
            })
        })

        Context("with invalid input", func() {
            It("should handle errors gracefully", func() {
                result, err := component.MethodName(ctx, "")

                Expect(err).To(HaveOccurred())
                Expect(result).To(BeEmpty())
            })
        })
    })
})
```

## Building

- Build all packages: `go build ./...`
- Build MCP server: `go build -o build/mcp-gds cmd/mcp/main.go`
- Check for Go formatting issues: `go fmt ./...`
- Check for common Go issues: `go vet ./...`

## Module Management

- Module name: `mgds`
- Go version: 1.23.0
