Commitizen
----------

CLI prompt tool for creation of [Conventional Commits](https://www.conventionalcommits.org). No more pain with commits rejected by validation tools or digging through [CONTRIBUTING.md](CONTRIBUTING.md) to find preferred commit format.

Based on wonderful work of [Commitizen](https://github.com/commitizen) with their [cz-cli](https://github.com/commitizen/cz-cli) which works like a charm in Node.js based projects.

This project replicated **cz-cli** way of work for Go projects

Installation
------------

With **Go** (>= 1.25):

```bash
go install github.com/isokolovskii/commitizen@1.0.0
```

-	or as a go tool

	```bash
	go get -tool github.com/isokolovskii/commitizen
	```

Usage
-----

Using the command line tool
---------------------------

TODO

As git hook with Lefthook
-------------------------

TODO

Authors and Contributors
------------------------

@isokolovskii (Ivan Sokolovskii, author)

Special thanks to @JimTheDev, whose [cz-cli](https://github.com/commitizen/cz-cli) project makes conventional commits easy to use in Node.js projects and was inspiration and background for my work on this Go adaptation of his work.
