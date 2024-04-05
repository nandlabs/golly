# Turbo
A webtool kit for golang

The `turbo` package provides enterprise grade http routing capabilities. The lightweight router works well
with all the necessary Use Cases and at the same time scales well.

---

- [Quick Start Guide](#quick-start-guide)
- [Features](#features)
  - [Base Routing](#base-routing)
  - [Multiple HTTP Methods Registering](#multiple-http-methods-registering)
  - [Routes Registering](#routes-registering)
  - [Path Params Wrapper](#path-params-wrapper)
  - [Query Params Wrapper](#query-params-wrapper)
  - [Filters](#filters)
- [Benchmarking Results](#benchmarking-results)

---




### Quick Start Guide

Being a Lightweight HTTP Router, it comes with a simple usage as explained below, just import the package, and you are
good to go.

```go
func main() {
    router := turbo.New()
    router.Get("/api/v1/healthCheck", healthCheck) // healthCheck is the handler Function
    router.Get("/api/v1/getData", getData) // getData is the handler Function

    srv := &http.Server{
        Handler:        router,
        Addr:           ":8080",
        ReadTimeout:    20 * time.Second,
        WriteTimeout:   20 * time.Second,
    }

    if err := srv.ListenAndServe(); err != nil {
        log.Fatalln(err)
    }
}
```

### Features

#### Base Routing

- Router lets you register routes based on the common HTTP Methods such as
    1. GET
       ```go
        router.Get("/api/v1/getCustomers", getCustomers) 
        ```
    2. POST
       ```go
        router.Post("/api/v1/getCustomers", getCustomers) 
        ```
    3. PUT
       ```go
        router.Put("/api/v1/getCustomers", getCustomers) 
        ```
    4. DELETE
        ```go
        router.Delete("/api/v1/getCustomers", getCustomers) 
        ```

#### Multiple HTTP Methods Registering

- Router lets you register routes with multiple methods such as `("POST", "PUT")` for a single endpoint.

  With the help of `Add` function that can be achieved
   ```go
    router.Add("/api/v1/addCustomers", getCustomers, "PUT", "POST") 
   ```
  This will register a route called `/api/v1/addCustomers` with two functions attached to a single route, `PUT`
  and `POST`

#### Routes Registering

- Routes can be registered in the following ways
    * Registering Static Routes
        ```go
        router.Get("/api/v1/getCustomers", getCustomers) 
        ```

    * Registering with Path Variables

      _The path variables can be registered with **:<name_of_param>**_
        ```go
        router.Get("/api/v1/getCustomer/:id", getCustomer)
        ```

#### Path Params Wrapper

- Path Params can be fetched with the built-in wrapper provided by the framework
    * The framework exposes a number of functions based on the type of variable that has been registered with the route
        * To fetch string parameters
            ```go
            getPathParms(id string, r *http.Request) string {}
            ```
        * To Fetch Int parameters
            ```go
            getIntPathParms(id string, r *http.Request) int {}
            ```
        * To fetch Float parameters
           ```go
           getFloatPathParms(id string, r *http.Request) float64 {}
           ```
        * To Fetch Boolean parameters
           ```go
           getBoolPathParms(id string, r *http.Request) bool {}
           ```

#### Query Params Wrapper

- Query Parameters can also be fetched with a built-in wrapper functions provided by the framework
    * The Framework exposes a number of wrapper functions which lets you fetch the query params of specific data type
      required
        * To fetch string query params
            ```go
            GetQueryParams(id string, r *http.Request) string {}
            ```
        * To Fetch Int query params
            ```go
            GetIntQueryParams(id string, r *http.Request) int {}
            ```
        * To fetch Float64 query params
           ```go
           GetFloatQueryParams(id string, r *http.Request) float64 {}
           ```
        * To Fetch Boolean query params
           ```go
           GetBoolQueryParams(id string, r *http.Request) bool {}
           ```

#### Filters

- Filters are available to add your custom middlewares to the `route`.

  Keeping in mind that all these middlewares/filters can be added at the route level only, this way giving you more
  freedom on how each route should behave in a microservice.

  `turbo` provides two main Filter Functions which can be leveraged easily and make your microservice more flexible
   ```go
   1. AddFilter(filters ...FilterFunc)
   2. AddAuthenticator(filter FilterFunc)
   ```
    * `AddFilter()`

      This Filter expects input of type `FilterFunc` i.e. `func(http.Handler) http.Handler`. You can declare your own
      filters of type FilterFunc as explained before and Add them to the AddFilter() as explained below
      ```go
      func main() {
        turboRouter := turbo.NewRouter()
        turboRouter.Get("/api/v1", ResponseHandler).AddFilter(loggingFilter, dummyFilter)
        
        srv := &http.Server{
            Handler: turboRouter,
            Addr:    ":9292",
        }
      }
      
      func loggingFilter(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            logger.Info("Filter Added")
            logger.Info(r.RequestURI)
            next.ServeHTTP(w, r)
            logger.Info("Filter Added again")
        })
      }
      
      func dummyFilter(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            logger.Info("Second Filter Added")
            next.ServeHTTP(w, r)
            logger.Info("Second Filter Added again")
        })
      }
      ```
    * `AddAuthenticator()`

      turbo is working on supporting all the major `Authentication` schemes in accordance with the OAS3 Specifications;

      | Authorization  | Status |
             | :---           | :----: |
      | Basic Auth     | WIP    |
      | JWT            | TBD    |
      | OAuth          | TBD    |
      | LDAP           | TBD    |

      An Authentication Filter can be implemented like below

      ```go
      func main() {
        turboRouter := turbo.NewRouter()
        var authenticator = auth.CreateBasicAuthAuthenticator()
        // provide the configuration to your authenticator filter struct, 
        // with the relevant struct Objects that would be exposed 
        // Once those input to your `authenticator` is fed, it can be used as the Filter easily
        turboRouter.Get("/api/v1", ResponseHandler).AddAuthenticator(authenticator)
        
        srv := &http.Server{
            Handler: turboRouter,
            Addr:    ":9292",
        }
      }
      ```

  `Working Understanding`

  The filters get executed in the order you add in the `AddFilter()` which states that if you add functions : f1, f2, f3
  as filters and want to be executed before the actual handler executes. The Order of execution chain becomes
    ```shell 
    f1 --> f2 --> f3 --> handlerFunction
    ```
  If you add Authentication Filter i.e. `AddAuthenticator()` explicitly and then add other filters to `AddFilter()`,
  then the order of execution of chain becomes,
    ```shell 
    authFilterFunc --> f1 --> f2 --> f3 --> handlerFunction
    ```
  Turbo gives the Authentication Filter precedence over any of the filter added to the chain. Rest all the chain order
  gets preserved in order they are added.

### Benchmarking Results

```bash
To be released soon
```
