
# CURL TEST

# register
curl -X POST http://localhost:<port>/register \


# login
curl -H "Authorization:<token>" http://localhost:<port>/login

# REST  ( UserService example )
curl -X GET http://localhost:<port>/users
curl -X GET http://localhost:<port>/users/messi
curl -X GET http://localhost:<port>/users?username=messi&age=18
curl -X GET http://localhost:<port>/users?username=messi&age=18




