Installing Summary
==========

        # for windows
        install MinGW. http://sourceforge.net/projects/tdm-gcc/
        install git. http://git-scm.com/download, add git and patch.exe to path
        install golang. http://golang.org/dl/
        install cmake. http://www.cmake.org/cmake/resources/software.html

        git clone https://github.com/funshine/heka
        cd heka
        source build.sh # Unix (or `. build.sh`; must be sourced to properly setup the environment)
        build.bat -DCMAKE_BUILD_TYPE=debug # Windows, run as Admin

        ctest             # All, see note
        # Or use the makefile target
        make test         # Unix
        mingw32-make test # Windows

        You will now have a ``hekad`` binary in the ``build/heka/bin`` directory.

        make install         # Unix
        mingw32-make install # Windows, run as Admin

Run 
===

        hekad -config=/path/to/test.toml

        C:\Program Files (x86)\heka\bin>hekad.exe -config="C:\Users\Hong\Projects\heka\externals\test.toml"

From Source
===========

`hekad` requires a Go work environment to be setup for the binary to be built;
this task is automated by the build process. The build script will override the
Go environment for the shell window it is executed in. This creates an isolated
environment that is intended specifically for building and developing Heka.
The build script should be sourced every time a new shell is opened for Heka
development to ensure the correct dependencies are found and being used. To
create a working `hekad` binary for your platform you'll need to install some
prerequisites. Many of these are standard on modern Unix distributions and all
are available for installation on Windows systems.

Prerequisites (all systems):

- CMake 3.0.0 or greater http://www.cmake.org/cmake/resources/software.html
- Git http://git-scm.com/download
- Go 1.4 or greater http://golang.org/dl/
- Mercurial http://mercurial.selenic.com/wiki/Download
- Protobuf 2.3 or greater (optional - only needed if message.proto is modified) http://code.google.com/p/protobuf/downloads/list
- Sphinx (optional - used to generate the documentation) http://sphinx-doc.org/
- An internet connection to fetch sub modules

Prerequisites (Unix):

- CA certificates (most probably already installed, via the ca-certificates package)
- make
- gcc and libc6 development headers (package glibc-devel or libc6-dev)
- patch
- GeoIP development files (optional)
- dpkg, debhelper and fakeroot (optional)
- rpmbuild (optional)
- packagemaker (optional)

Prerequisites (Windows):

- MinGW http://sourceforge.net/projects/tdm-gcc/


Build Instructions
------------------

1. Check out the `heka` repository:

        git clone https://github.com/mozilla-services/heka

2. Source (Unix-y) or run (Windows) the build script in the heka directory:

        cd heka
        source build.sh # Unix (or `. build.sh`; must be sourced to properly setup the environment)
        build.bat  # Windows

You will now have a ``hekad`` binary in the ``build/heka/bin`` directory.

3. (Optional) Run the tests to ensure a functioning `hekad`:

        ctest             # All, see note
        # Or use the makefile target
        make test         # Unix
        mingw32-make test # Windows


    In addition to the standard test build target, ctest can be called directly
    providing much greater control over the tests being run and the generated
    output (see ctest --help). i.e., 'ctest -R pi' will only run the pipeline
    unit test.

4. Run ``make install`` to install libs and modules into a usable location:

       make install         # Unix
       mingw32-make install # Windows

   This will install all of Heka's required support libraries, modules, and
   other files into a usable ``share_dir``, at the following path:

       /path/to/heka/repo/heka/share/heka

5. Specify Heka configuration:

   When setting up your Heka configuration, you'll want to make sure you
   set the global ``share_dir`` setting to point to the path above. The
   ``[hekad]`` section might look like this:

       [hekad]
       maxprocs = 4
       share_dir = "/path/to/heka/repo/heka/share/heka"

Clean Targets
-------------
- clean-heka - Use this target any time you change branches or pull from the Heka repository, it will ensure the Go workspace is in sync with the repository tree.
- clean - You will never want to use this target (it is autogenerated by cmake), it will cause all external dependencies to be re-fetched and re-built.  The best way to 'clean-all' is to delete the build directory and re-run the build.(sh|bat) script.


Build Options
-------------

There are two build customization options that can be specified during the cmake generation process.

- INCLUDE_MOZSVC (bool) Include the Mozilla services plugins (default Unix: true, Windows: false).
- BENCHMARK (bool) Enable the benchmark tests (default false)

For example: to enable the benchmark tests in addition to the standard unit tests
upon building type 'source ./build.sh -DBENCHMARK=true ..' in the top repo directory.


Building `hekad` with External Plugins
======================================

It is possible to extend `hekad` by writing input, decoder, filter, or output
plugins in Go (see :ref:`plugins`). Because Go only supports static linking of
Go code, your plugins must be included with and registered into Heka at
compile time. The build process supports this through the use of an optional 
cmake file `{heka root}/cmake/plugin_loader.cmake`.  A cmake function has been
provided `add_external_plugin` taking the repository type (git, svn, or hg), 
repository URL, the repository tag to fetch, and an optional list of 
sub-packages to be initialized.


        add_external_plugin(git https://github.com/mozilla-services/heka-mozsvc-plugins 6fe574dbd32a21f5d5583608a9d2339925edd2a7)
        add_external_plugin(git https://github.com/example/path <tag> util filepath)
        add_external_plugin(git https://github.com/bellycard/heka-sns-input :local)
        # The ':local' tag is a special case, it copies {heka root}/externals/{plugin_name} into the Go 
        # work environment every time `make` is run. When local development is complete, and the source
        # is checked in, the value can simply be changed to the correct tag to make it 'live'.
        # i.e. {heka root}/externals/heka-sns-input -> {heka root}/build/heka/src/github.com/bellycard/heka-sns-input

The preceding entry clones the `heka-mozsvc-plugins` git repository into the Go
work environment, checks out SHA 6fe574dbd32a21f5d5583608a9d2339925edd2a7, and imports the package into 
`hekad` when `make` is run. By adding an `init() function <http://golang.org/doc/effective_go.html#init>`_ 
in your package you can make calls into `pipeline.RegisterPlugin` to register 
your plugins with Heka's configuration system.


Creating Packages
=================

Installing packages on a system is generally the easiest way to deploy
`hekad`. These packages can be easily created after following the above
:ref:`From Source <from_source>` directions:

1. Run `cpack` to build the appropriate package(s) for the current
system:

        cpack                # All
        # Or use the makefile target
        make package         # Unix (no deb, see below)
        make deb             # Unix (if dpkg is available see below)
        mingw32-make package # Windows

The packages will be created in the build directory.

    You will need `rpmbuild` installed to build the rpms.

    seealso:: `Setting up an rpm-build environment <http://wiki.centos.org/HowTos/SetupRpmBuildEnvironment>`_


    For file name convention reasons, deb packages won't be created by running
    `cpack` or `make package`, even on a Unix machine w/ dpkg installed.
    Instead, running `source build.sh` on such a machine will generate a
    Makefile with a separate 'deb' target, so you can run `make deb` to
    generate the appropriate deb package. Additionnaly, you can add a suffix to
    the package version, for example:


        CPACK_DEBIAN_PACKAGE_VERSION_SUFFIX=+deb8 make deb
