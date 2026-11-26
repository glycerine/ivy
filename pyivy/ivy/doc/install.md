---
layout: page
title: Installing IVy
---

There are two ways to install ivy:

1. [Install from source](#source)
2. [Install a wheel](#binary)

<a name="source"></a> Installing from source
--------------------------------------------

1. [Install from source on Linux](#linuxnotes)
2. [Install from source on macOS](#macnotes)


<a name="linuxnotes"></a> Installation from source on Linux
===========================================================

These commands use Python 3.10 or later on Ubuntu and other
Debian-based distributions.

### Prerequisites

    $ sudo apt-get install python3 python3-pip python3-venv g++ cmake git python3-tk graphviz pkg-config libssl-dev libreadline-dev

### Install IVy

Use the `pyivy/ivy` directory of the GoIvy checkout.

    $ cd ~/ivy/pyivy/ivy
    $ python3 -m venv .venv
    $ . .venv/bin/activate
    $ python build_submodules.py
    $ python -m pip install .

For development, install the package in editable mode with
`python -m pip install -e .`. PyIvy uses Ken McMillan's vendored Z3 4.7.1
fork for its interpolation support. The build script prepares that
library along with the native tools used for model checking.

Pass `--with-picotls` to build and install the TLS libraries. After
building native dependencies, reinstall PyIvy to include them in a
non-editable installation.

Optionally, build the experimental Ivy v2.0 compiler:

    $ python build_v2_compiler.py

### Run

Run Ivy on an example, like this:

    $ cd doc/examples
    $ ivy client_server_example.ivy

Or, if you only want to use Ivy on the command line, test it like this:

    $ ivy_check trace=true doc/examples/client_server_example_new.ivy
    
Ivy should print out a counterexample trace.

### Emacs mode

An emacs major mode for Ivy is available in `lib/emacs/ivy-mode.el`. Put this file
somewhere in your emacs load path and add the following code to your
`.emacs`:

    (add-to-list 'auto-mode-alist '("\\.ivy\\'" . ivy-mode))
    (autoload 'ivy-mode  "ivy-mode.el" "Major mode for editing Ivy code" t nil)

    
<a name="macnotes"></a> Installation from source on macOS
=======================================================

Install the Xcode command-line tools and the Homebrew prerequisites.
Follow [Homebrew's installation instructions](https://brew.sh/) to
install it and add it to your path.

    $ xcode-select --install
    $ brew install cmake coreutils git graphviz pkgconf openssl@3 readline python@3.10 python-tk@3.10

Create the environment with `python3.10 -m venv .venv` and follow the
[source installation](#linuxnotes) above. The native build script
builds ABC and AIGER on macOS. Set
`OPENSSL_ROOT_DIR` to use a custom OpenSSL for the optional picotls build.
The C++ extractor links the same vendored Z3 fork.

<a name="binary"></a> Installing a wheel
--------------------------------------

Install a wheel built from this checkout into an active virtual
environment. The examples and documentation are available in the checkout.

    $ python -m pip install /path/to/ms_ivy-<version>-<platform>.whl

The upstream IVy project publishes its own `ms-ivy` wheels on PyPI.
