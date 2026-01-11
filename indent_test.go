package pretty

import (
	"testing"
)

func TestIndent(t *testing.T) {
	tests := []struct {
		name       string
		source     string
		indent     string
		linePrefix []string
		want       string
	}{
		{
			name:   "empty string",
			source: "",
			indent: "  ",
			want:   "",
		},
		{
			name:   "simple struct no indent",
			source: `Struct{X:1;Y:2}`,
			indent: "  ",
			want: `Struct{
  X: 1
  Y: 2
}`,
		},
		{
			name:   "nested struct",
			source: `Struct{X:1;Sub:{Y:2}}`,
			indent: "  ",
			want: `Struct{
  X: 1
  Sub: {
    Y: 2
  }
}`,
		},
		{
			name:   "empty braces",
			source: `Struct{X:{}}`,
			indent: "  ",
			want: `Struct{
  X: {}
}`,
		},
		{
			name:   "multiple empty braces",
			source: `{X:{};Y:{}}`,
			indent: "  ",
			want: `{
  X: {}
  Y: {}
}`,
		},
		{
			name:   "raw string with special chars",
			source: "Struct{Str:`hello:world;{}`}",
			indent: "  ",
			want: `Struct{
  Str: ` + "`hello:world;{}`" + `
}`,
		},
		{
			name:   "escaped string with special chars",
			source: `Struct{Str:"hello:world;{}"}`,
			indent: "  ",
			want: `Struct{
  Str: "hello:world;{}"
}`,
		},
		{
			name:   "escaped string with backslash",
			source: `Struct{Path:"C:\\Users\\file"}`,
			indent: "  ",
			want: `Struct{
  Path: "C:\\Users\\file"
}`,
		},
		{
			name:   "escaped string with escaped quote",
			source: `Struct{Str:"say \"hello\""}`,
			indent: "  ",
			want: `Struct{
  Str: "say \"hello\""
}`,
		},
		{
			name:   "colon adds space",
			source: `{X:1;Y:2}`,
			indent: "  ",
			want: `{
  X: 1
  Y: 2
}`,
		},
		{
			name:   "semicolon adds newline",
			source: `{X:1;Y:2;Z:3}`,
			indent: "  ",
			want: `{
  X: 1
  Y: 2
  Z: 3
}`,
		},
		{
			name:   "tab indent",
			source: `{X:1;Y:2}`,
			indent: "\t",
			want: `{
	X: 1
	Y: 2
}`,
		},
		{
			name:       "with line prefix",
			source:     `{X:1;Y:2}`,
			indent:     "  ",
			linePrefix: []string{"// "},
			want: `// {
//   X: 1
//   Y: 2
// }`,
		},
		{
			name:       "with multiple line prefixes",
			source:     `{X:1}`,
			indent:     "  ",
			linePrefix: []string{"  ", "// "},
			want: `  // {
  //   X: 1
  // }`,
		},
		{
			name:   "deeply nested",
			source: `{A:{B:{C:{D:1}}}}`,
			indent: "  ",
			want: `{
  A: {
    B: {
      C: {
        D: 1
      }
    }
  }
}`,
		},
		{
			name:   "unbalanced braces - more close",
			source: `{X:1}}`,
			indent: "  ",
			want: `{
  X: 1
}
}`,
		},
		{
			name:   "unbalanced braces - more open",
			source: `{{X:1}`,
			indent: "  ",
			want: `{
  {
    X: 1
  }`,
		},
		{
			name:   "UTF-8 content",
			source: `{Name:"🎉";Value:"hello 世界"}`,
			indent: "  ",
			want: `{
  Name: "🎉"
  Value: "hello 世界"
}`,
		},
		{
			name:   "no trailing content after last brace",
			source: `{X:1;Y:2}`,
			indent: "  ",
			want: `{
  X: 1
  Y: 2
}`,
		},
		{
			name:   "content after closing brace",
			source: `{X:1}.Field`,
			indent: "  ",
			want: `{
  X: 1
}.Field`,
		},
		{
			name:   "mixed raw and escaped strings",
			source: "{Raw:`{};`;Esc:\"{};\";}",
			indent: "  ",
			want:   "{\n  Raw: `{};`\n  Esc: \"{};\"\n  \n}",
		},
		{
			name:   "backtick in escaped string",
			source: "{Str:\"hello`world\"}",
			indent: "  ",
			want: `{
  Str: "hello` + "`" + `world"
}`,
		},
		{
			name:   "quote in raw string",
			source: "{Str:`hello\"world`}",
			indent: "  ",
			want: `{
  Str: ` + "`" + `hello"world` + "`" + `
}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Indent([]byte(tt.source), tt.indent, tt.linePrefix...)
			if string(got) != tt.want {
				t.Errorf("Indent() = %q, want %q", string(got), tt.want)
			}
		})
	}
}

func TestIndent_EdgeCases(t *testing.T) {
	t.Run("nil source", func(t *testing.T) {
		result := Indent(nil, "  ")
		if result == nil {
			t.Error("Indent() returned nil, want non-nil slice")
		}
		if len(result) != 0 {
			t.Errorf("Indent() = %q, want empty slice", string(result))
		}
	})

	t.Run("empty indent string", func(t *testing.T) {
		result := Indent([]byte("{X:1;Y:2}"), "")
		expected := `{
X: 1
Y: 2
}`
		if string(result) != expected {
			t.Errorf("Indent() = %q, want %q", string(result), expected)
		}
	})

	t.Run("single character", func(t *testing.T) {
		result := Indent([]byte("{"), "  ")
		expected := `{
  `
		if string(result) != expected {
			t.Errorf("Indent() = %q, want %q", string(result), expected)
		}
	})

	t.Run("only semicolons", func(t *testing.T) {
		result := Indent([]byte(";;;"), "  ")
		expected := `


`
		if string(result) != expected {
			t.Errorf("Indent() = %q, want %q", string(result), expected)
		}
	})

	t.Run("only colons", func(t *testing.T) {
		result := Indent([]byte(":::"), "  ")
		expected := ": : : "
		if string(result) != expected {
			t.Errorf("Indent() = %q, want %q", string(result), expected)
		}
	})

	t.Run("unclosed raw string", func(t *testing.T) {
		result := Indent([]byte("{Str:`hello"), "  ")
		expected := `{
  Str: ` + "`hello"
		if string(result) != expected {
			t.Errorf("Indent() = %q, want %q", string(result), expected)
		}
	})

	t.Run("unclosed escaped string", func(t *testing.T) {
		result := Indent([]byte(`{Str:"hello`), "  ")
		expected := `{
  Str: "hello`
		if string(result) != expected {
			t.Errorf("Indent() = %q, want %q", string(result), expected)
		}
	})

	t.Run("backslash at end of escaped string", func(t *testing.T) {
		result := Indent([]byte(`{Str:"hello\`), "  ")
		expected := `{
  Str: "hello\`
		if string(result) != expected {
			t.Errorf("Indent() = %q, want %q", string(result), expected)
		}
	})

	t.Run("consecutive empty braces", func(t *testing.T) {
		result := Indent([]byte("{}{}{}"), "  ")
		expected := "{}{}{}"
		if string(result) != expected {
			t.Errorf("Indent() = %q, want %q", string(result), expected)
		}
	})

	t.Run("empty braces in nested structure", func(t *testing.T) {
		result := Indent([]byte("{A:{};B:{}}"), "  ")
		expected := `{
  A: {}
  B: {}
}`
		if string(result) != expected {
			t.Errorf("Indent() = %q, want %q", string(result), expected)
		}
	})
}

func TestIndent_RealWorldExamples(t *testing.T) {
	t.Run("Sprint output from struct", func(t *testing.T) {
		// This is typical output from Sprint
		source := `Struct{Parent{Map:nil};Int:0;Str:"hello";Sub:{Map:{}}}`
		result := Indent([]byte(source), "  ")
		expected := `Struct{
  Parent{
    Map: nil
  }
  Int: 0
  Str: "hello"
  Sub: {
    Map: {}
  }
}`
		if string(result) != expected {
			t.Errorf("Indent() = %q, want %q", string(result), expected)
		}
	})

	t.Run("Sprint output with nested maps", func(t *testing.T) {
		source := `{Map:{Key1:Value1;Key2:Value2};Slice:[1,2,3]}`
		result := Indent([]byte(source), "  ")
		expected := `{
  Map: {
    Key1: Value1
    Key2: Value2
  }
  Slice: [1,2,3]
}`
		if string(result) != expected {
			t.Errorf("Indent() = %q, want %q", string(result), expected)
		}
	})

	t.Run("Error struct formatting", func(t *testing.T) {
		source := `ErrorStruct{X:1;Y:2;err:"something failed"}`
		result := Indent([]byte(source), "  ")
		expected := `ErrorStruct{
  X: 1
  Y: 2
  err: "something failed"
}`
		if string(result) != expected {
			t.Errorf("Indent() = %q, want %q", string(result), expected)
		}
	})

	t.Run("multiple nested structs", func(t *testing.T) {
		source := `Root{First:{A:1;B:2};Second:{C:3;D:4};Third:{E:{F:5;G:6}}}`
		result := Indent([]byte(source), "  ")
		expected := `Root{
  First: {
    A: 1
    B: 2
  }
  Second: {
    C: 3
    D: 4
  }
  Third: {
    E: {
      F: 5
      G: 6
    }
  }
}`
		if string(result) != expected {
			t.Errorf("Indent() = %q, want %q", string(result), expected)
		}
	})

	t.Run("struct with map", func(t *testing.T) {
		source := `Struct{Name:"test";Config:{timeout:30;retries:3;debug:true}}`
		result := Indent([]byte(source), "  ")
		expected := `Struct{
  Name: "test"
  Config: {
    timeout: 30
    retries: 3
    debug: true
  }
}`
		if string(result) != expected {
			t.Errorf("Indent() = %q, want %q", string(result), expected)
		}
	})

	t.Run("struct with short slice", func(t *testing.T) {
		source := `Struct{IDs:[1,2,3];Names:["a","b","c"]}`
		result := Indent([]byte(source), "  ")
		expected := `Struct{
  IDs: [1,2,3]
  Names: ["a","b","c"]
}`
		if string(result) != expected {
			t.Errorf("Indent() = %q, want %q", string(result), expected)
		}
	})

	t.Run("struct with longer slice", func(t *testing.T) {
		source := `Struct{Items:[{ID:1;Name:"first"},{ID:2;Name:"second"},{ID:3;Name:"third"}]}`
		result := Indent([]byte(source), "  ")
		expected := `Struct{
  Items: [{
    ID: 1
    Name: "first"
  },{
    ID: 2
    Name: "second"
  },{
    ID: 3
    Name: "third"
  }]
}`
		if string(result) != expected {
			t.Errorf("Indent() = %q, want %q", string(result), expected)
		}
	})

	t.Run("complex nested with map, structs and slices", func(t *testing.T) {
		source := `Response{Status:200;Data:{Users:[{ID:1;Name:"Alice";Roles:["admin","user"]},{ID:2;Name:"Bob";Roles:["user"]}];Meta:{Total:2;Page:1}}}`
		result := Indent([]byte(source), "  ")
		expected := `Response{
  Status: 200
  Data: {
    Users: [{
      ID: 1
      Name: "Alice"
      Roles: ["admin","user"]
    },{
      ID: 2
      Name: "Bob"
      Roles: ["user"]
    }]
    Meta: {
      Total: 2
      Page: 1
    }
  }
}`
		if string(result) != expected {
			t.Errorf("Indent() = %q, want %q", string(result), expected)
		}
	})
}

func BenchmarkIndent(b *testing.B) {
	source := []byte(`Struct{Parent{Map:nil};Int:0;Str:"hello world";Sub:{Map:{Key1:Value1;Key2:Value2;Key3:Value3}}}`)
	indent := "  "

	b.ResetTimer()
	for range b.N {
		_ = Indent(source, indent)
	}
}

func BenchmarkIndent_DeepNesting(b *testing.B) {
	source := []byte(`{A:{B:{C:{D:{E:{F:{G:{H:{I:{J:value}}}}}}}}}}`)
	indent := "  "

	b.ResetTimer()
	for range b.N {
		_ = Indent(source, indent)
	}
}

func BenchmarkIndent_LongString(b *testing.B) {
	source := []byte(`{Field1:1;Field2:2;Field3:3;Field4:4;Field5:5;Field6:6;Field7:7;Field8:8;Field9:9;Field10:10}`)
	indent := "  "

	b.ResetTimer()
	for range b.N {
		_ = Indent(source, indent)
	}
}
