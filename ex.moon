# ------------------------------------------------------------------------------
# example config format
#
# this is a working draft of things that are valid in a new config language to
# replace json as a config language for Go projects.
#
# comments are a thing now!
# ------------------------------------------------------------------------------

# the whole document is implicitly a namespace, so you can set key value pairs
# at the top level.
first_name: jordan
last_name: orelli

# the bare strings true and false are boolean values
bool_true: true
bool_false: false

# the quoted strings "true" and "false" are string values.  In the unlikely
# event you need literal true and false strings, quote them.
string_true: "true"
string_false: "false"

# lists of things should be supported
items: [
    one
    2
    3.4
    [five; 6 7.8]
]

# objects should be supported
hash: {key: value; other_key: other_value}

other_hash: {
    key_1: one
    key_2: 2
    key_3: 3.4
    key_4: [five; 6 7.8]
}

# we may reference an item that was defined earlier using a sigil
repeat_hash: @hash

# items can be hidden.  i.e., they are only valid in the parse and eval stage
# as intermediate values internal to the config file; they are *not* visible to
# the host program.  This is generally useful for composing larger, more
# complicated things.
@hidden_item: it has a value
visible_item: @hidden_item

@person_one: {
    name: the first name here
    age: 28
    hometown: crooklyn
}

@person_two: {
    name: the second name here
    age: 30
    hometown: tha bronx
}

people: [@person_one @person_two]
