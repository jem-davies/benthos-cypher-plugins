# benthos-cypher-plugins
Benthos Plugins for Graph DB engines that use the Cypher query language - Neo4j, Memgraph etc.

The current status of this Repo and it's plugins are a work in progress. 


## How to use

You will need to run Benthos using go run, rather than calling the Benthos binary

Create a new go project and add the below to your main.go:

```go
package main

import (
	"context"

	_ "github.com/benthosdev/benthos/v4/public/components/all" // (you don't have to import all components)
	"github.com/benthosdev/benthos/v4/public/service"
	_ "github.com/jem-davies/benthos-cypher-plugins" // import this repo
)

func main() {
	service.RunCLI(context.Background())
}
```

This should then register the plugins as components and you can add them to your pipeline config.


## Example Cypher Output: 

```yaml
input:
  file:
    paths: ["./input.json"]
    codec: all-bytes

output: 
  cypher:
    Database: "neo4j" 
    Uri: "bolt://localhost:7687"
    User: "neo4j"
    Password: "password" #Auth has not yet been implemented - you must use a neo4j DB that has Auth disabled in it's settings
```

Where input.json contains: 

```json
{
    "SOR": "Subject,SubjectType,Object,ObjectType,Relation\nRemus,PERSON,Romulus,PERSON,brother\nRomulus,PERSON,Remus,PERSON,brother\nRemus,PERSON,Rome,CITY,founded_by\nRomulus,PERSON,Rome,CITY,founded_by"
}
```

This should then produce the following graph in Neo4j: 


        FOUNDED_BY       ┌───────┐
    ┌────────────────────┤ROMULUS│
    │                    └─▲──┬──┘
    │                      │  │
    │                     B│  │B
 ┌──▼─┐                   R│  │R
 │ROME│                   O│  │O
 └──▲─┘                   .│  │.
    │                      │  │
    │                      │  │
    │   FOUNDED_BY       ┌─┴──▼─┐
    └────────────────────┤REMUS │
                         └──────┘



