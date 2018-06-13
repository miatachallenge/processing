# MiataChallenge Processing

A web application that runs a pipeline transforming information written into
a RethinkDB database by Sync to proper laps, filtering them using scripts,
attributing them to a driver using a Google Spreadsheet and eventually saving
them to a CouchDB server using a format renderably by [mc.nickesh.pl](http://mc.nickesh.pl).

![Screenshot of the UI](https://i.imgur.com/2RjLqGG.png)

The program accepts a `-rdb_addr` flag, which is the address and port of the
RethinkDB server (the `mctracker` database is used) and `-db_path` is a path
to the directory where a JSON flatfile database will be created for storing
all the parameters. The server listens to `:8070` by default.

`convert` and `validate` should be JS functions - the first one converting
the tag ID into the value of the EPC column (without spaces) from the
spreadsheet, and `validate` returning a boolean signifying whether a lap time
of `arg1` milliseconds is valid. You should tune up `validate` before each race,
using sensible defaults.

The "Upstream address" should point to a CouchDB database. The credentials
should be specified in the URL if they are required, using the standard HTTP
Basic Auth format. The "Drivers spreadsheet" should be an ID of a Google
Spreadsheet to use for the drivers mapping. In order to use the spreadsheet,
you must generate a Service Account in the Google API Management site and save it
in the working directory as `client_secret.json`.

Run the application using either Docker or directly as a normal Go app. Remember
to keep the `frontend` directory in the workdir!
