// Copyright (C) MongoDB, Inc. 2017-present.
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at http://www.apache.org/licenses/LICENSE-2.0

// NOTE: This documentation should be kept in line with the Example* test functions.

// Package mongo provides a MongoDB Driver API for Go.
//
// Basic usage of the driver starts with creating a Client from a connection
// string. To do so, call the NewClient and Connect functions:
//
// 		client, err := NewClient(options.Client().ApplyURI("mongodb://foo:bar@localhost:27017"))
// 		if err != nil { return err }
// 		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
// 		defer cancel()
// 		err = client.Connect(ctx)
// 		if err != nil { return err }
//
// This will create a new client and start monitoring the MongoDB server on localhost.
// The Database and Collection types can be used to access the database:
//
//    collection := client.Database("baz").Collection("qux")
//
// A Collection can be used to query the database or insert documents:
//
//    res, err := collection.InsertOne(context.Background(), bson.M{"hello": "world"})
//    if err != nil { return err }
//    id := res.InsertedID
//
// Several methods return a cursor, which can be used like this:
//
//    cur, err := collection.Find(context.Background(), bson.D{})
//    if err != nil { log.Fatal(err) }
//    defer cur.Close(context.Background())
//    for cur.Next(context.Background()) {
//      // To decode into a struct, use cursor.Decode()
//      result := struct{
//        Foo string
//        Bar int32
//      }{}
//      err := cur.Decode(&result)
//      if err != nil { log.Fatal(err) }
//      // do something with result...
//
//      // To get the raw bson bytes use cursor.Current
//      raw := cur.Current
//      // do something with raw...
//    }
//    if err := cur.Err(); err != nil {
//      return err
//    }
//
// Methods that only return a single document will return a *SingleResult, which works
// like a *sql.Row:
//
// 	  result := struct{
// 	  	Foo string
// 	  	Bar int32
// 	  }{}
//    filter := bson.D{{"hello", "world"}}
//    err := collection.FindOne(context.Background(), filter).Decode(&result)
//    if err != nil { return err }
//    // do something with result...
//
// All Client, Collection, and Database methods that take parameters of type interface{}
// will return ErrNilDocument if nil is passed in for an interface{}.
//
// Additional examples can be found under the examples directory in the driver's repository and
// on the MongoDB website.
//
// Potential DNS Issues
//
// Building with Go 1.11+ and using connection strings with the "mongodb+srv"[1] scheme is
// incompatible with some DNS servers in the wild due to the change introduced in
// https://github.com/golang/go/issues/10622. If you receive an error with the message "cannot
// unmarshal DNS message" while running an operation, we suggest you use a different DNS server.
//
// Client Side Encryption
//
// Client-side encryption is a new feature in MongoDB 4.2 that allows specific data fields to be encrypted.
//
// Important: This feature is beta. The API for both automatic and explicit encryption/decryption does not have any
// stability guarantees and backwards-breaking changes may be made before the final release.
//
// Note: Auto encryption is an enterprise-only feature.
//
// The libmongocrypt C library is required when using client-side encryption. To install libmongocrypt, do the following:
//
//    // run the clone command in an empty folder because the compile script will create new directories.
//    git clone https://github.com/mongodb/libmongocrypt
//    ./libmongocrypt/.evergreen/compile.sh
//
//    // Linux/Darwin: (this requires pkg-config to be installed on the system)
//    Add <installation_dir>/install/libmongocrypt/lib/pkgconfig:<installation_dir>/install/mongo-c-driver/lib/pkgconfig to PKG_CONFIG_PATH.
//    Add <installation_dir>/install/libmongocrypt/lib to LD_LIBRARY_PATH
//
//    // Windows:
//    mkdir -p c:/libmongocrypt/include
//    mkdir -p c:/libmongocrypt/bin
//    cp ./install/libmongocrypt/lib/mongocrypt.dll c:/libmongocrypt/bin
//    cp ./install/libmongocrypt/include/mongocrypt/*.h c:/libmongocrypt/include
//    // add c:/libmongocrypt/bin to PATH
//
// libmongocrypt communicates with the mongocryptd process for automatic encryption. This process can be started manually
// or auto-spawned by the driver itself. To enable auto-spawning, ensure the process binary is on the PATH. To start it
// manually, use AutoEncryptionOptions:
//
//    aeo := options.AutoEncryption()
//    mongocryptdOpts := map[string]interface{}{
//        "mongocryptdBypassSpawn": true,
//    }
//    aeo.SetExtraOptions(mongocryptdOpts)
// To specify a process URI for mongocryptd, the "mongocryptdURI" option can be passed in the ExtraOptions map as well.
// More information about mongocryptd will soon be available from the official documentation.
//
// [1] See https://docs.mongodb.com/manual/reference/connection-string/#dns-seedlist-connection-format
package mongo
