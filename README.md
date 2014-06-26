# redisgosearch

redisgosearch implements fast full-text search with Golang and Redis, using Redis's rich support for sets.

This fork of @sunfmin's original allows custom keyword segmentation, and cleaned-up interfaces with documentation. The tests also have no external dependencies other than a working Redis installation.

The [original documentation](https://theplant.jp/en/blogs/13-techforce-making-a-simple-full-text-search-with-golang-and-redis) is clarified below.

## Tutorial

Check out [godoc](http://godoc.org/github.com/purohit/redisgosearch) for package documentation.

Let's say you have blog entries:

```go
type Entry struct {
    Id          string
    Title       string
    Content     string
}
```

You want to be able to search on the `Title` and `Content` fields. Let's create two blog entries to index.

```go
Entry {
    Id:      "50344415ff3a8aa694000001",
    Title:   "Organizing Go code",
    Content: "Go code is organized differently from that of other languages. This post discusses",
}

Entry {
    Id:      "50344415ff3a8aa694000002",
    Title:   "Getting to know the Go community",
    Content: "Over the past couple of years Go has attracted a lot of users and contributors",
}
```

All keys in Redis are prefixed by a namespace you pass in when creating a `Client` (in this case, `entries`).

When you call `Index` on the two entries, the text from `Title` and `Keyword` is broken up into keywords by the `DefaultSegment` function (if you have your own keyword segmentation function, call `IndexCustom`) Each key's value is a set whose members point back to the original entries.

```
redis 127.0.0.1:6379> keys *
1) "entries:keywords:go"
2) "entries:keywords:community"
3) ...

redis 127.0.0.1:6379> SMEMBERS entries:keywords:go
1) "entries:entity:50344415ff3a8aa694000001"
2) "entries:entity:50344415ff3a8aa694000002"

redis 127.0.0.1:6379> SMEMBERS entries:keywords:community
1) "entries:entity:50344415ff3a8aa694000002"
```

That way, searching for entries that belong to a query such as "go community" is a simple set intersection. The query is first segmented to `["go", "community"]`, and the intersection (Redis: `SINTER`) is performed to return the IDs of the original entities.

```
redis 127.0.0.1:6379> SINTER entries:keywords:go entries:keywords:community
1) "entries:entity:50344415ff3a8aa694000002"
```
Then, `Search` will take the resulting keys, unmarshal the original structs,
and return them to you. redisgosearch can index any Go struct satisfying `Indexable`.

```go
type Entry struct {
    Id          string
    GroupId     string
    Title       string
    Content     string
    Attachments []*Attachment
}

func (entry *Entry) IndexPieces() (r []string, ais []redisgosearch.Indexable) {
    r = append(r, entry.Title)
    r = append(r, entry.Content)

    for _, a := range entry.Attachments {
        r = append(r, a.Filename)
        ais = append(ais, &IndexedAttachment{entry, a})
    }

    return
}

func (entry *Entry) IndexEntity() (indexType string, key string, entity interface{}) {
    key = entry.Id
    indexType = "entries"
    entity = entry
    return
}

func (entry *Entry) IndexFilters() (r map[string]string) {
    r = make(map[string]string)
    r["group"] = entry.GroupId
    return
}
```

`IndexPieces` tells the package what text should be segmented and indexed. In our example, you might also want to index other data connected to an entry, like attachment data, so you could search any filename and find out which entries those files belong to. Thus, `ais` can return an array of `Indexable` objects that are indexed and connected with the original struct.

`IndexEntity` tells the package the string indexType (used to prefix keys), and the unique key. Combined with the namespace, this becomes the key that owns a Redis `SET`. The actual entity struct will be marshalled into JSON and stored into Redis.

`IndexFilters` allows metadata to further filter queries. For example, because we added a filter above, you can search “go community” filtered by the "group" “New York”:

```go
var entries []*Entry
count, err := client.Search("entries", "go community",
                map[string]string{"group": "New York"},
                0, 20, &entries)
```

The 0 and 20 is for pagination, and `count` is the total number of entries that matched "go community".

## Contributing
The current feature set is simple, and new features are appreciated. Please initiate a pull request, and make sure to `go fmt` and `golint`!
