package query

// Example usage of the configuration system:
//
// 1. Using default configuration:
//    graph, err := BuildGraphHybrid(tree, sqlitePath, badgerPath, nil)
//
// 2. Using custom configuration:
//    config := DefaultConfig()
//    config.Cache.HybridNodeCacheSize = 100000
//    config.Cache.HybridXrefCacheSize = 50000
//    config.Timeout.SQLiteQueryTimeout = 60 * time.Second
//    graph, err := BuildGraphHybrid(tree, sqlitePath, badgerPath, config)
//
// 3. Loading from config file:
//    config, err := LoadConfig("") // Searches standard locations
//    if err != nil {
//        log.Fatal(err)
//    }
//    graph, err := BuildGraphHybrid(tree, sqlitePath, badgerPath, config)
//
// 4. Loading from specific config file:
//    config, err := LoadConfig("/path/to/config.json")
//    if err != nil {
//        log.Fatal(err)
//    }
//    graph, err := BuildGraphHybrid(tree, sqlitePath, badgerPath, config)
//
// 5. Saving configuration to file:
//    config := DefaultConfig()
//    config.Cache.HybridNodeCacheSize = 75000
//    if err := SaveConfig(config, ""); err != nil {
//        log.Fatal(err)
//    }

