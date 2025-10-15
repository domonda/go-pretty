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
}
```

Set to `0` or negative values to disable truncation.

## License

See LICENSE file
