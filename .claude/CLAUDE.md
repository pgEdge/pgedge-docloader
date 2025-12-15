* Always use 4 spaces for indentation of code.

* Always update documentation where necessary:

    * When editing markdown files, always leave a blank line before the first 
        item in any list or sub-list to ensure the lists render properly in 
        tools such as mkdocs.
    * /README.md is a brief summary for users looking at the code:
    * All documentation content must be available in the documentation under 
        /docs.
    * Markdown files in the root directory should use uppercase names (except 
        for the extension). 
    * Markdown files in /docs should have lowercase names.
    * The documentation engine is MKDocs, which is configured in mkdocs.yml in
        the root of the project. The TOC needs to be maintained whenever the 
        documentation content in /docs is updated.
    * Documentation markdown files MUST be wrapped at 79 characters or less.

* Always add tests to exercise new functionality.
    * When running tests to verify changes, always run all tests and check 
        verbose output for failures or errors.
    * Don't tail or otherwise trim test output to both stdout and stderr when 
        running tests, to ensure nothing is missed.
    * Don't modify any tests unless they are expected to fail as a result of 
        the changes being made.
    * Ensure any temporary files created during test runs are removed when the 
        run is complete, except for logs that may need to be reviewed.
    * Ensure all tests run under the "go test" suite.
    * ALWAYS use the top-level Makefile for running tests, e.g. make lint or 
        make test

* Remember to ensure that all code changes remain secure:
    * Always escape inputs from the user to protect against injection attacks.
    * Ensure defensive coding techniques are always employed.
