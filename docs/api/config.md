## Example

Query whole configuration:

```graphql
query {
    config {
        global {
            tproxy_port
            log_level
            tcp_check_url
            udp_check_dns
            check_interval
            check_tolerance
            lan_interface
            wan_interface
            allow_insecure
            dial_mode
        }
        group {
            name
            filter {
                ...FragmentAndFunctions
            }
            policy {
                ...FragmentAndFunctionsOrPlaintext
            }
        }
        dns {
            upstream {
                key
                value
            }
            routing {
                request {
                    ...FragmentRouting
                }
                response {
                    ...FragmentRouting
                }
            }
        }
        routing {
            ...FragmentRouting
        }
    }
}

fragment FragmentFunction on Function {
    name
    params {
        key
        value
    }
}

fragment FragmentAndFunctions on AndFunctions {
    and {
        ...FragmentFunction
    }
}

fragment FragmentAndFunctionsOrPlaintext on AndFunctionsOrPlaintext {
    __typename
    ...on Plaintext {
        value
    }
    ...on AndFunctions {
        ...FragmentAndFunctions
    }
}

fragment FragmentFunctionOrPlaintext on FunctionOrPlaintext {
    __typename
    ...on Function {
        ...FragmentFunction
    }
    ...on Plaintext {
        value
    }
}

fragment FragmentRouting on Routing {
    rules {
        conditions {
            ...FragmentAndFunctions
        }
    }
    fallback {
        ...FragmentFunctionOrPlaintext
    }
}
```
