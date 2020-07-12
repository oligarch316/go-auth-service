#!/usr/bin/make -f

# Disable all default suffixes
.SUFFIXES:

# ----- Variables
AUTHSVC_DB_DATA_DIR ?= data

sqlite_db_path := $(AUTHSVC_DB_DATA_DIR)/sqlite.db

# ----- Default
.PHONY: default

default: sqlite

# ----- SQLite
.PHONY: sqlite sqlite.debug sqlite.clean

$(sqlite_db_path): sqlite-init.sql
	$(info Initializing sqlite database)
	@mkdir -p $(@D)
	@sqlite3 -init $< $@ '.exit'

sqlite: $(sqlite_db_path)

sqlite.debug: $(sqlite_db_path)
	$(info Opening sqlite interpreter)
	@sqlite3 $<

sqlite.clean:
	$(info Cleaning sqlite database)
	@rm -f $(sqlite_db_path)

# ----- Mongo
.PHONY: mongo mongo.debug mongo.clean

mongo:
	$(error Mongo not yet implemented)

mongo.debug:
	$(error Mongo not yet implemented)

mongo.clean:
	$(error Mongo not yet implemented)

# ----- Clean
.PHONY: clean

clean: sqlite.clean

# ----- HELP
.PHONY: help

print-%: ; @echo "$($*)"

# TODO: help
