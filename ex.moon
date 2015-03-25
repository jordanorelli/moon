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
first_name: "jordan"
last_name: "orelli"

# lists of things should be supported
items: [
    "one"
    2
    3.4
    ["five" 6 7.8]
]

# objects should be supported
hash: {key: "value" other_key: "other_value"}

other_hash: {
    key_1: "one"
    key_2: 2
    key_3: 3.4
    key_4: ["five" 6 7.8]
}

item_one: "this is item one"

# we should be able to reference variables defined earlier
# (this doesn't work yet)
# item_two: item_one
