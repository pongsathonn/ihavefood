TODO (DONE) handle rabbitmq consistant every services 
TODO (DONE) add log for getEnv every services
TODO (DONE) UserService refactor code , create more data layer for repository

# PRIORITY SORT

TODO auth , tested assign role in middleware ( rold invalid ) correct (maybe)
TODO group init to a function
TODO add role in claims (JWT)
TODO modify userService , authservice follow food.proto
TODO might modify some setupMux middleware for role
TODO add Role-Based Access Control (RBAC)
    FLOW
    my app contains 3 roles "visitor" | "user" | "admin"
    - visitor can access register endpoint we treat this as "visitor"
    - new user registerd (default role "user")
    - rootuser create when app start has privilage to give user role "admin" | "user"
    - admin role can access restrict api such as "DELETE /api/users/foobar"
    - optional = add permission to role i.e admin role can "read" | "write" , user role can only "read"

TODO add debug mode for compose (optional)
TODO validate new placeorder that this user is exists
TODO might remove pb from user service ( repository should be single response not incluse protobuff )
TODO delivery service clear everyfunction logic
TODO finished every function in each service , espeacially Delivery service ( deos not implement any algorithm )
TODO coupon service implement everyfuncion and add check coupon in placeorder ( gateway )
TODO check format email and phonenumber
TODO change dockerfile cmd from "docker-auth" to somethingelse
TODO handle environment variable in compose file and each service 
TODO manage rabbitmqctl see which exchange do , and which service bind queue etc , might be use ack when send and receive
TODO curl test every endpoints ( look at food.proto )
TODO Dockerfile ight use multi stage build
TODO handler error godly 
TODO write UML correctly

TODO add Unittest (optional)
TODO Deploy

TODO learn more about context in GO ( for experienced )
TODO practice algorithm


TODO Find a Job !!!!!!!!!!!!

