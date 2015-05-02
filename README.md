This is a toy configuration language.  If you want to know why, see the [WHY](WHY.md) doc.

# Instability

This isn't even remotely stable.  Don't use this in an actual project, I
promise you I will break this.

# What it is

Moon is a configuration language intended to replace JSON as a configuration
format in Go projects.  The syntax is fairly similar, with a few changes:

- instead of having some hand-wavy "number" thing without ever specifying the
  valid ranges or anything like that, there are two number types: int and
  float64.  Int means int32 on 32bit systems and int64 on 64bit systems.  Maybe
  I'll go back on that because it sounds really dumb when I type it out.
- support for complex numbers.  The only reason I included this is that it's
  included in the text/template package in the standard library and I stole the
  number-parsing code from that library.
- bare strings.  Support here is somewhat up in the air.  Inside of bare
  strings, the characters [, ], ;, :, {, and } must be escaped, because each of
  those can terminate a bare string.
- comments.  There are no comments in JSON and that's incredibly annoying.
  Moon has them.
- variables.  In moon, you can make a reference to a value that was already
  defined.
- hidden variables.  This sounds really dumb but it quickly became apparent
  that if you're using variables to compose something, those variables end up
  polluting the namespace of the moon document.
- a command-line tool.  I'm sure these exist for json but there's no official
  "this is the tool, that one isn't" decree.
- something kinda like xpath that lets you select elements in a document.

# How it works

json and similar formats parse text and render some kind of parse tree.  We do
the same thing in moon, but then we walk the tree and evaluate it, which is why
variables are a thing.  It's effectively just enough of a dynamic language
interpreter for variables to work.

# Types

Moon defines the following types:

- integers: right now this is an int based on Go semantics; it's a 32 bit int
  on 32 bit CPUs, and a 64 bit int on 64 bit CPUs.  These are some integers:

```
1
2
-1
-12348
0
+0
```

- floats: they're all float64.  These are some floats:

```
1.0
1.2
-9.3
3.14
1e9
```

- complex numbers: they're complex128 values in Go.  These are some complex numbers:

```
1+2i
-9+4i
```

- strings: they're strings.  They're not explicitly required to be composed of
  UTF-8 runes but I haven't really been testing binary data, so for the moment,
  all bets are off here.  They're quoted, but maybe I'll go back on that.
  These are strings:

```
this is a bare string
"this is a quoted string"
'this is also a quoted string'

inside of a bare string, "quotes" don't need to be escaped
but semicolons \;, colons \:, parens \( and \), brackets \[ and \] and braces \{ \} need to be escaped.
```

  You can use single or double quotes.  Escape quotes with a backslash.  Quoted
  strings may contain newlines and special characters.

  The following characters are special characters: `:`, `;`, `[`, `]`, `#`,
  `{`, and `}`.  The colon is used to separate a key from a value.  The
  semicolon is used to terminate a bare string.  A newline will also terminate
  a bare string.  A close bracket must be a special string in order to support
  a bare string being the last element of a list, without requiring that you
  add a semicolon after it.  An open bracket doesn't need to be a special
  character, but I made it a special character to be symetrical with the close
  bracket.  The same logic applies for braces.

- objects: or maybe you call them hashes, objects, or associative arrays.  Moon
  calls them objects, but you'd never know it because it's actually the
  `map[string]interface{}` type in Go, which is effectively the same thing.
  Keys are bare strings, but object keys may not contain spaces.

  These are some objects:

```
{name: "jordan" age: 28}
{
    one: 1
    two: two is also a number
    pi: 3.14
}
```

- lists: they're `[]interface{}` values.  They're not typed, and they can be
  heterogenous.  Values are separated by spaces.  I might put commas back in,
  that's kinda up in the air right now.  These are some lists:

```
[1 2 3]
[
  one
  2
  3.14
]

# this is a list of three elements
[moe; larry; curly]

# this is a list of one element
[moe larry curl]
```

- variables: a variable is indicated with a `@` character.  A variable may
  refer to any key that has already been defined in the current moon document.
  Here is an example of how you would use a variable:

```
original_value: this is the original value, which is a string
duplicate_value: @original_value
```

  If the name of a key begins with an `@` character, it is a private key.  A
  private key may be referenced in the rest of the moon document, but it is not
  available anywhere outside of the moon document parser.  This is useful for
  the composition of larger, nested values, without polluting the document's
  root namespace.  Here is an example of a private key:

```
@my_private_key: a great value
my_public_key: @my_private_key
```

  This isn't particularly useful for simple values, but consider the following:

```
@prod: {
    label: production
    hostname: prod.example.com
    port: 9000
    user: prod-user
}

@dev: {
    label: development
    hostname: dev.example.com
    port: 9200
    user: dev-user
}

servers: [@prod @dev]
```

  The only key in the root namespace of the document is `servers`; the
  individual server configs are considered only a part of the document.  Since
  we know that these values are private to the document itself, we know we are
  free to modify them as we see fit, without affecting how the host program
  sees the moon document.
