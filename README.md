# Fitness API

## REST API
POST    /exercise/                  : create a exercise entry, returns ID
GET     /exercise/<exerciseid>      : returns a single exercise by ID
GET     /exercise/                  : returns all exercises
DELETE  /exercise/<exerciseid>      : delete an exercise by ID
GET     /exercise/<yy>/<mm>/<dd>    : returns list of exercises submitted on this date
GET     /tag/<tagname>              : returns list of exercises with this tag

