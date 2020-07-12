#!/usr/bin/make -f

# Disable all default suffixes
.SUFFIXES:

# ----- Variables
prefix := /usr/local
exec_prefix := $(prefix)
bindir := $(exec_prefix)/bin

INSTALL = install
INSTALL_PROGRAM = $(INSTALL) -m 755

VERSION ?= $(shell git describe --always)
git_revision := $(shell git rev-parse HEAD)

go_version_package := $(shell go list ./pkg/version)
go_sources := $(shell find . -type f -name '*.go')

binary_tail := _$(shell go env GOOS)_$(shell go env GOARCH)
binary_targets := $(shell find cmd -maxdepth 1 -mindepth 1 -type d -exec basename {} \;)
binary_targets_static := $(binary_targets:%=%.static)

install_targets := $(binary_targets:%=%.install)
install_targets_static := $(binary_targets_static:%=%.install)

link_targets := $(install_targets:%.install=%.link)
link_targets_static := $(install_targets_static:%.install=%.link)

build_flags :=
link_vars := appVersion=$(VERSION) gitRevision=$(git_revision)
ld_flags := -ldflags '$(link_vars:%=-X "$(go_version_package).%")'

# ----- Aliases
.PHONY: default install link static static.install static.link

default: authsvc
install: authsvc.install
link: authsvc.link

static: authsvc.static
static.install: authsvc.static.install
static.link: authsvc.static.link

# ----- Build
.PHONY: $(binary_targets) $(binary_targets_static)

define dynamic_binary_target
bin/$(1)%: $(go_sources) go.mod go.sum
	$$(info Compiling binary '$$(notdir $$@)')
	@mkdir -p $$(@D)
	@go build $$(build_flags) $$(ld_flags) -o $$@ cmd/$(1)/main.go
endef

$(foreach name,$(binary_targets),$(eval $(call dynamic_binary_target,$(name))))

$(binary_targets): %: bin/%$(binary_tail)

$(binary_targets_static): export CGO_ENVABLED=0
$(binary_targets_static): build_flags += -a -installsuffix cgo
$(binary_targets_static): %.static: bin/%$(binary_tail)_static


# ----- Install
.PHONY: $(install_targets) $(install_targets_static) $(link_targets) $(link_targets_static)

$(DESTDIR)$(bindir)/%:
	$(info Installing binary '$*' via '$(INSTALL_PROGRAM)')
	@$(INSTALL) -d $(@D)
	@$(INSTALL_PROGRAM) $(abspath bin/$*$(binary_tail)) $@

$(install_targets): %.install: $(DESTDIR)$(bindir)/%

$(install_targets_static): binary_tail := $(binary_tail)_static
$(install_targets_static): %.static.install: %.install

$(link_targets) $(link_targets_static): INSTALL_PROGRAM = ln -s
$(link_targets) $(link_targets_static): %.link: %.install

# ----- Tools
.PHONY: fmt lint tidy

fmt:
	$(info Foramatting)
	@go fmt ./...

lint:
	$(info Linting)
	@golint ./...

tidy:
	$(info Tidying go modules)
	@go mod tidy

# ----- Clean
.PHONY: clean

clean:
	$(info Cleaning binaries)
	@rm -rf bin

# ----- HELP
.PHONY: help

print-%: ; @echo "$($*)"

# TODO: help
