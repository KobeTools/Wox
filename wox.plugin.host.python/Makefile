.PHONY: clean build install help

HOSTS_DIR := ../wox.core/resource/hosts
SDK_DIR := ../wox.plugin.python

# GNU Make can run this subdirectory build directly from Windows shells where
# Git usr/bin is missing, so expose the POSIX helpers used by rm here instead of
# relying on the root Makefile to have normalized PATH first.
ifeq ($(OS),Windows_NT)
    GIT_USR_BIN := $(patsubst %/bin/sh.exe,%/usr/bin,$(SHELL))
    ifneq ($(GIT_USR_BIN),$(SHELL))
        export PATH := $(GIT_USR_BIN);$(PATH)
    endif
endif

help:
	@echo "Available commands:"
	@echo "  make clean      - Remove build artifacts and cache files"
	@echo "  make build      - Build Python host"
	@echo "  make install    - Install development dependencies"
	@echo "  make lint       - Run linting"
	@echo "  make format     - Run formatting"

clean:
	rm -rf python-host.pyz python-host/ $(SITE_DIR)

install:
	uv cache clean
	uv sync --all-extras --locked

lint:
	uv run ruff check src
	uv run mypy src

format:
	uv run ruff format src

test:
	uv run python -m unittest discover -s tests -v

# KobeTools fork: build the zipapp from uv.lock with hash checks. Upstream let
# shiv run pip against PyPI (unpinned, bypassing uv.lock). Third-party deps are
# installed --require-hashes from the lock; the host and SDK are pure Python and
# are copied from source, so no build backend is fetched. `format` is no longer
# part of the build (it rewrote sources).
SITE_DIR := .build-site

build: clean install lint
	rm -rf $(SITE_DIR) && mkdir -p $(SITE_DIR) $(HOSTS_DIR)
	uv export --frozen --no-dev --no-emit-project --no-emit-package wox-plugin --format requirements-txt -o $(SITE_DIR)/requirements.lock
	uv pip install --target $(SITE_DIR)/site --require-hashes --no-deps -r $(SITE_DIR)/requirements.lock
	cp -R src/wox_plugin_host $(SDK_DIR)/src/wox_plugin $(SITE_DIR)/site/
	uv run shiv --compressed --reproducible --site-packages $(SITE_DIR)/site -e wox_plugin_host.__main__:run -o $(HOSTS_DIR)/python-host.pyz
	rm -rf $(SITE_DIR)
