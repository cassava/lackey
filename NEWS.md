Lackey Releases
================

## Version 0.8 (9 January 2023)
This release turns the Docker image into one that uses [gotty](https://github.com/sorenisanerd/gotty)
to provide a service by default. This provides a simple web server
to view the previous or current progress of a job and start a new one.

A port is opened on 8080. You can completely override this by overriding
the default command that is run.

## Version 0.7 (8 January 2023)
This release adds a Dockerfile and the options to downscale an album
cover when copying it. See options: `--downscale-cover`, `--cover-source`,
and `--cover-target`.

## Version 0.6 (7 February 2021)
This release adds the `--copy-suffix` option which allows you to force some
audio files to be copied instead of transcoded.

## Version 0.4
This release adds OPUS encoding support.

## Version 0.3 (12 January 2017)
This release changes the versioning scheme to use semantic versioning.
It is also the first release where I start to pay more attention to
everything in general.

Since we are still changing a lot of program details and functionality, we are
in an unstable state, hence the major version number being 0. We chose 3 as the
minor number because that was the next one coming.

The primary change from the previous version is improved documentation
and the use of dependency vendoring.
