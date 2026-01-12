# go-pretty

Pretty printing for complex Go types

## Overview

`go-pretty` provides compact, single-line pretty printing of Go values suitable for logging and debugging. It handles circular references, implements smart truncation, and offers customizable formatting options.

## Features

- **Single-line output**: All values formatted on one line for easy log parsing
- **Circular reference detection**: Automatically detects and handles circular data structures
- **Smart string handling**: Byte slices containing valid UTF-8 are printed as strings
- **Truncation support**: Configurable limits for strings, errors, and slices
- **Custom formatting**: Implement `Printable` interface for custom types
- **Null handling**: Support for nullable types via `Nullable` interface
- **Sorted maps**: Map keys are automatically sorted for consistent output
- **Context awareness**: Special handling for `context.Context` types

## Installation

```bash
go get github.com/domonda/go-pretty
```

## Usage

### Basic Printing

```go
import "github.com/domonda/go-pretty"

// Print to stdout
pretty.Println(myStruct)
pretty.Print(myValue)

// Print to string
s := pretty.Sprint(myValue)

// Print to io.Writer
pretty.Fprint(writer, myValue)
pretty.Fprintln(writer, myValue)
```

### Indented Output

All print functions accept optional indent arguments:
- **No arguments**: prints on a single line without indentation
- **One argument**: uses `indent[0]` as indent string for nested structures
- **Two+ arguments**: uses `indent[0]` as indent string and `indent[1:]` concatenated as line prefix

```go
// Single line output (no indentation)
pretty.Println(myStruct)

// Multi-line with 2-space indentation
pretty.Println(myStruct, "  ")

// Multi-line with 2-space indentation and line prefix ">>> "
pretty.Println(myStruct, "  ", ">>> ")

// Multi-line with line prefix concatenated from multiple strings
pretty.Println(myStruct, "  ", "[", "LOG", "] ")  // Line prefix: "[LOG] "
```

### Custom Printer Configuration

```go
printer := pretty.Printer{
    MaxStringLength: 200,  // Truncate strings longer than 200 chars
    MaxErrorLength:  2000, // Truncate errors longer than 2000 chars
    MaxSliceLength:  20,   // Truncate slices with more than 20 elements
}

printer.Println(myValue)
s := printer.Sprint(myValue)
```

### JSON Output

```go
// Print as indented JSON
pretty.PrintAsJSON(myStruct)

// Custom indent
pretty.PrintAsJSON(myStruct, "    ")
```

### Custom Type Formatting

Implement the `Printable` interface for custom formatting:

```go
type MyType struct {
    field string
}

func (m MyType) PrettyPrint(w io.Writer) {
    fmt.Fprintf(w, "MyType{%s}", m.field)
}
```

### Nullable Types

Implement the `Nullable` interface to print "null" for zero values:

```go
type MyNullable struct {
    value *string
}

func (m MyNullable) IsNull() bool {
    return m.value == nil
}
```

### Advanced: Custom Formatting with PrintFuncFor

The `Printer.PrintFuncFor` field allows you to customize how values are printed based on their `reflect.Value`. This is useful when you want to:
- Add custom formatting for types you don't control
- Adapt types that implement different interfaces (e.g., `fmt.Stringer`, custom serializers)
- Change formatting based on runtime conditions
- Wrap values with additional context

#### Example: Adapting Other Interfaces

You can use `PrintFuncFor` to enable pretty printing for types that implement other interfaces:

```go
import (
    "fmt"
    "io"
    "reflect"
    "github.com/domonda/go-pretty"
)

// Assume you have types implementing fmt.Stringer or custom interfaces
type CustomStringer struct {
    Name string
}

func (c CustomStringer) String() string {
    return fmt.Sprintf("Custom<%s>", c.Name)
}

// Create a printer that handles fmt.Stringer types
printer := pretty.DefaultPrinter.WithPrintFuncFor(func(v reflect.Value) pretty.PrintFunc {
    stringer, ok := v.Interface().(fmt.Stringer)
    if !ok && v.CanAddr() {
        stringer, ok = v.Addr().Interface().(fmt.Stringer)
    }
    if ok {
        return func(w io.Writer) {
            fmt.Fprint(w, stringer.String())
        }
    }
    return pretty.PrintFuncForPrintable(v) // Use default
})

printer.Println(CustomStringer{Name: "test"})
// Output: Custom<test>
```

#### Example: Runtime Conditional Formatting

```go
// Mask sensitive data based on type or field tags
printer := pretty.DefaultPrinter.WithPrintFuncFor(func(v reflect.Value) pretty.PrintFunc {
    // Customize based on type name
    if v.Kind() == reflect.String && v.String() == "a sensitive string" {
        return func(w io.Writer) {
            fmt.Fprint(w, "`***REDACTED***`")
        }
    }
    return pretty.PrintFuncForPrintable(v) // Use default
})

printer.Println("a sensitive string")
// Output: `***REDACTED***`
```

**Note:** If `Printer.PrintFuncFor` is not set, the `PrintFuncForPrintable` function is used, which checks if the value implements the `Printable` interface.

### Integration with go-errs

The `go-errs` package uses a configurable `Printer` variable (of type `*pretty.Printer`) for formatting function parameters in error call stacks. You can customize this printer to mask secrets, adapt types, or change formatting without implementing the `Printable` interface on your types.

**Use cases:**
- Hide sensitive data (secrets, passwords, tokens) in error messages and stack traces
- Customize call stack formatting in error output
- Mask PII (Personally Identifiable Information) in logs
- Adapt types that implement other interfaces globally

```go
import (
    "fmt"
    "io"
    "reflect"
    "strings"
    "github.com/domonda/go-errs"
    "github.com/domonda/go-pretty"
)

func init() {
    // Configure the Printer used by go-errs for error call stacks
    errs.Printer.PrintFuncFor = func(v reflect.Value) pretty.PrintFunc {
        // Mask sensitive strings
        if v.Kind() == reflect.String {
            str := v.String()
            // Check for common secret patterns
            if strings.Contains(str, "password") ||
               strings.Contains(str, "token") ||
               strings.Contains(str, "secret") {
                return func(w io.Writer) {
                    fmt.Fprint(w, "`***REDACTED***`")
                }
            }
        }

        // Hide sensitive struct fields
        if v.Kind() == reflect.Struct {
            t := v.Type()
            for i := 0; i < t.NumField(); i++ {
                field := t.Field(i)
                // Check for "secret" tag
                if field.Tag.Get("secret") == "true" {
                    // Return custom formatter that masks this field
                    // (implementation would format all fields except sensitive ones)
                }
            }
        }

        return pretty.PrintFuncForPrintable(v) // Use default
    }
}

// Now all error stack traces from go-errs will automatically mask secrets
// without needing to implement Printable on your types
```

This approach allows you to:
1. **Centrally control** how all values are formatted in error call stacks
2. **Protect sensitive data** in logs, error traces, and debug output
3. **Customize error formatting** without modifying go-errs code
4. **Apply formatting rules** to types you don't control

**Note:** The `errs.Printer` variable is a `*pretty.Printer` that can be fully configured with custom settings like `MaxStringLength`, `MaxErrorLength`, `MaxSliceLength`, and `PrintFuncFor`.

## Output Examples

```go
// Strings are backtick-quoted
pretty.Sprint("hello")  // `hello`

// Structs show field names
type Person struct { Name string; Age int }
pretty.Sprint(Person{"Alice", 30})  // Person{Name:`Alice`;Age:30}

// Slices and arrays
pretty.Sprint([]int{1, 2, 3})  // [1,2,3]

// Maps with sorted keys
pretty.Sprint(map[string]int{"b": 2, "a": 1})  // map[string]int{`a`:1;`b`:2}

// Circular references
type Node struct { Next *Node }
n := &Node{}
n.Next = n
pretty.Sprint(n)  // Node{Next:CIRCULAR_REF}

// Byte slices as strings
pretty.Sprint([]byte("hello"))  // `hello`

// Time and Duration
pretty.Sprint(time.Now())  // Time(`2024-01-15T10:30:00Z`)
pretty.Sprint(5*time.Second)  // Duration(`5s`)

// Nil values
pretty.Sprint((*int)(nil))  // nil
pretty.Sprint(error(nil))  // nil

// Context
ctx := context.Background()
pretty.Sprint(ctx)  // Context{}
```

## Configuration

The default printer used by package-level functions:

```go
var DefaultPrinter = Printer{
    MaxStringLength: 200,
    MaxErrorLength:  2000,
    MaxSliceLength:  20,
    PrintFuncFor:    nil,
}
```

Set to `0` or negative values to disable truncation.

## License

See LICENSE file
