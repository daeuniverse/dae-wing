# [dae](https://github.com/daeuniverse/dae)-wing

## Run

To run the api only:

```bash
make deps
go run . run -c ./ --api-only
# go build -o dae-wing && ./dae-wing run -c ./ --api-only
```

To run with dae:

```bash
make deps
go run -exec sudo . run
# go build -o dae-wing && sudo ./dae-wing run -c ./ --api-only
```

## API

API is powered by [GraphQL](https://graphql.org/). UI developers can export schema and write queries and mutations easily.

```bash
git clone https://github.com/daeuniverse/dae-wing
go build -o dae-wing
./dae-wing export schema > schema.graphql
```

[graphql-playground](https://github.com/graphql/graphql-playground) is recommended for developers. It integrates docs and debug environment for API. Choose `URL ENDPOINT` and fill in `http://localhost:2023/graphql` to continue.

### Config generator

Alternatively, you can use [raw format inputs](https://github.com/daeuniverse/dae/blob/main/example.dae), use [dae-outline2config](https://github.com/daeuniverse/dae-outline2config) to generate config related raw format.

To generate outline:

```bash
git clone https://github.com/daeuniverse/dae-wing
go build -o dae-wing .
./dae-wing export outline > outline.json
```

## Structure

### Config

Config defined in dae-wing includes `global`, `dns` and `routing` sections in [dae](https://github.com/daeuniverse/dae).

Users can switch between multiple configs. Nodes, subscriptions and groups are selectively shared by all configs.

**Run**

Selected config is the running config or config to run. If dae is not running, you can select a config and invoke run. If dae is already running with a config, selecting a new config will cause automatic switching and reloading, and removing the running config will cause to stop running.

### Subscription

Subscription consists of its link and the collection of nodes resolved by the link.

### Node

A generalized node refer to a proxy profile, which can be imported by link. A node can be in a subscription or not. It depends on how it is imported. Nodes in the same collection must have unique links, which means nodes will be deduplicated by dae-wing before being added to a collection.

### Group

A group has the following features:

- A group is as an outbound of routing.
- A group consists of subscriptions, nodes and a node selection policy for every new connection.

If a node in a subscription also belongs to a group, it will be preserved when the subscription is updated.
