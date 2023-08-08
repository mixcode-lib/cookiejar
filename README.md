
# Modified Go cookiejar with JSON marshalling

This package is essentially the same codebase with Go's net/http/cookiejar, but with marshalling functions added.


## added functions

There are 4 functions added.

* **MarshalJson()** and **MarshalJsonIndent()** : Make a JSON corresponding to the cookiejar's state.

* **MergeJson()**: Merge the cookies in the marshalled JSON into the jar.

* **Clear()**: clear the jar.

