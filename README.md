This is a toy configuration language.  If you want to know why, see the [WHY](WHY.md) doc.

# stability

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
- comments.  There are no comments in JSON and that's incredibly annoying.
  Moon has them.
- variables.  In moon, you can make a reference to a value that was already
  defined.
- hidden variables.  This sounds really dumb but it quickly became apparent
  that if you're using variables to compose something, those variables end up
  polluting the namespace of the moon document.  Variables whose name starts
  with a period are hidden; they are only valid in the context of the moon
  document in which they appear, and they are not accessible to the host
  program.
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

    1  
    2  
    -1  
    -12348  
    0  
    +0  

- floats: they're all float64.  These are some floats:

    1.0  
    1.2  
    -9.3  
    3.14  
    1e9  

- complex numbers: they're complex128 values in Go.  These are some complex numbers:

    1+2i  
    -9+4i  

- strings: they're strings.  They're not explicitly required to be composed of
  UTF-8 runes but I haven't really been testing binary data, so for the moment,
  all bets are off here.  They're quoted, but maybe I'll go back on that.
  These are strings:

    "this one"  
    'that one'  

  You can use single or double quotes.  Escape quotes with a backslash.
- objects: or maybe you call them hashes, objects, or associative arrays.  Moon
  calls them objects, but you'd never know it because it's actually the
  `map[string]interface{}` type in Go, which is effectively the same thing.
  Unlike json objects, and unlike strings, the keys in objects are not quoted.
  Also there are no commas between the values but I'm not sure I like this yet,
  so it might change.  These are some objects:

    {name: "jordan" age: 28}  
    {  
        one: 1  
        two: "two is also a number"  
        pi: 3.14  
    }  

- lists: they're `[]interface{}` values.  They're not typed, and they can be
  heterogenous.  Values are separated by spaces.  I might put commas back in,
  that's kinda up in the air right now.  These are some lists:

    [1 2 3]  
    [  
      "one"  
      2  
      3.14  
    ]  



