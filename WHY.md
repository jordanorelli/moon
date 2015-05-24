### Why?

Quite often in the course of writing Go programs I've found myself in need of a
format to describe the application's configuration.  The only formats in the Go
standard library are json and xml.

#### Why not XML?

xml is fairly verbose and doesn't map to data structures particularly well.
There are many ways of expressing the same thing.  E.g., what's the difference
between this:

```
<person>
    <first_name>jordan</first_name>
    <last_name>orelli</last_name>
</person>
```

and this:  `<person first_name="jordan" last_name="orelli" />`?

There are also spurious type indicators.  What do you do if you just want an
array of integers?

```
<sizes>
    <size>1</size>
    <size>3</size>
    <size>8</size>
</sizes>
```

In reality, all I want for that is something like `sizes: [1 3 8]`.

On top of the combination ambiguity and verbosity of the language itself, the
xml package in the Go standard library is cumbersome to work with.  The few
times I've tried it, it didn't strike me as being a particularly pleasurable
programming model.  It's not particularly well-suited to the task of writing
the host application, and it's not particularly well-suited to the task of
writing the config files themselves.

#### Why not JSON?

Historically, I have always used json as the configuration format for Go
projects.  I've found this to be the best option.  It's in the standard
library, so it doesn't add any dependencies to your project.  For small
configurations, json is very easy for a human to write.  It's a nearly ideal
format, inasmuch as it doesn't litter itself with spurious information, it's
relatively compact, and it's a familiar set of non-word characters that
everyone can grok.  The parsing API is also quite nice, and the json package in
the standard library is extremely easy to use.  As far as programming the host
application, json works very well.

But over time, using json gets worse and worse.  When it comes to authoring
json config files, things are a little less rosey.  At first it seems fine,
when your configs are small and your projects are simple.  Over time, as the
projects get more complex and you require more configuration, the files become
challenging to read and edit.  It's exceptionally easy to introduce syntax
errors into json documents.  Object keys seem to require double-quotes for no
reason at all.  There are no comments.  You're more or less pidgeonholed into
making your whole configuration file one object, because if you don't, you wind
up complicating the parsing API for the host application, and it becomes
difficult to program, which was one of the best features of using json to begin
with.  Then there are the weird things, like the fact that re-declaring a key
in an object just overwrites the previous value
([ex](http://play.golang.org/p/ky4F9UmM1Z)), which fundamentally makes no
sense because objects are supposed to be order-independent but in reality
they're not.  As far as performing automated validations, json doesn't give you
much help.

It also seems to drive my colleagues in ops into an absolute rage.

#### Why not YAML?

Semantic whitespace makes it difficult to generate YAML from within a template
language, such as ERB.  Generating configuration files from templates or from a
parent application is a common strategy when using configuration management
systems, such as [Chef](https://www.chef.io/chef/).

