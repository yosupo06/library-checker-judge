import Leaf
import Vapor
import PostgreSQL

/// Called before your application initializes.
public func configure(_ config: inout Config, _ env: inout Environment, _ services: inout Services) throws {    
    // Register providers first
    try services.register(PostgreSQLProvider())
    try services.register(LeafProvider())

    // Connect PostgreSQL
    let postgresql = PostgreSQLDatabase(config: PostgreSQLDatabaseConfig(
        hostname: Environment.get("POSTGRE_HOST") ?? "127.0.0.1",
        port: Int(Environment.get("POSTGRE_PORT") ?? "5432")!,
        username: Environment.get("POSTGRE_USER") ?? "postgres",
        database: "librarychecker",
        password: Environment.get("POSTGRE_PASS") ?? "passwd"
    ))
    var databases = DatabasesConfig()
    databases.add(database: postgresql, as: .psql)
    services.register(databases)

    // Connection pool
    services.register(DatabaseConnectionPoolConfig(maxConnections: 4))

    // Register routes to the router
    let router = EngineRouter.default()
    try routes(router)
    services.register(router, as: Router.self)
    
    // Use Leaf for rendering views
    config.prefer(LeafRenderer.self, for: ViewRenderer.self)

    // Register middleware
    var middlewares = MiddlewareConfig() // Create _empty_ middleware config
    middlewares.use(FileMiddleware.self) // Serves files from `Public/` directory
    middlewares.use(ErrorMiddleware.self) // Catches errors and converts to HTTP response
    services.register(middlewares)
}
