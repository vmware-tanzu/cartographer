# Meta
[meta]: #meta
- Name: Unit Testing: Onion Skin approach
- Start Date: 2022-07-14
- Author(s): Rasheed Abdul-Aziz <rabdulaziz@vmware.com>
- Supersedes: "N/A"

# Summary
[summary]: #summary

Using an onion-skin style of unit-testing, where business object testing stacks on top of more
specific business objects, rather than imposing test boundaries at the function or struct boundaries.

# Definitions
[definitions]: #definitions

* Public Boundary - the publicly accessible methods, variables and constants declared in code.
* Deterministic - Given the same inputs, the outputs will always be the same.
    * Specific exception for `Time.Now` which behaves as a near constant in the span of a unit test's operation.

# Motivation
[motivation]: #motivation


Our goal is to produce a prescriptive approach to writing unit tests that:

## Reduce DI
Reduce the use of DI for the sake of tests. This approach ensures that the public interface of objects
represents intentional modelling for the sake of API consumers, and not for the sake of a test.

## Pragmatic
Choosing to inject becomes intention revealing, but choosing not to inject when little value is introduced
is supported.

eg: Injecting a deterministic timer fake vs `time ~= now`

## Guiding principles for choosing boundaries
Make the boundaries of a 'unit' concise and easy to articulate through guiding principles.

We feel that this last point is important if we hope to see unit tests maintained by open source
contributors. There are many ways to select the boundaries for a test, so the prescription herein
provides a way to minimize bouncing of PRs due to stylistic choices.

# What it is
[what-it-is]: #what-it-is

This provides a high level overview of the feature.

- Define any new terminology.
- Define the target persona: buildpack author, buildpack user, platform operator, platform implementor, and/or project contributor.
- Explaining the feature largely in terms of examples.
- If applicable, provide sample error messages, deprecation warnings, or migration guidance.
- If applicable, describe the differences between teaching this to existing users and new users.

# How it Works
[how-it-works]: #how-it-works

The guiding architectural choices of this kind of testing involve:

* Lift IO toward the controller layer, away from business modelling, such as
  the [repoitory pattern](https://martinfowler.com/eaaCatalog/repository.html)
* Creating core business modelling around 'determinism'. Determinism occurs when
  code does not involve outside systems, including network and disk.

### Deterministic boundary testing

* Any code that accepts inputs (which include receiver structs) and produces deterministic
  output can be tested without stubbing.
    * **Without** defaulting to stubbing out determenistic dependencies.
* Any dependency on non-deterministic behavior should be stubbed and emulated, either with a
  deterministic stand-in or a [counterfeiter](https://github.com/maxbrunsfeld/counterfeiter) fake.
    * These should be Dependency injected for unit testing, not via variable replacement.

**Rule of thumbs**:

* Start with only async/non-deterministic dependencies stubbed out
* Use determenism and accept the duplication of test inputs up and down the stack (onion skin tests)

### When to break the "rules of thumb"

Only injecting non-deterministic dependencies means that a unit test will grow in complexity and will show layers
of duplicate test conditions. That's OK until:

* There is an obvious business domain that can be extracted. Rule of thumb: try to make the SUT look more 'generic',
  extracting the business domain logic as a dependency
    * This leads to dependencies that minimize business domain dependencies to a business domain layer.
    * TBD: show examples
* The test is so complex that it's time to search for the aforementioned business domain.

**Rule of thumb**:

* Let the tests guide you to when it might be time to extract interfaces/functions/structs into new,
  less dependant units.

## But my more concrete tests look like integrations!?

Those tests will:

* ensure the refactor is still functioning as expected
* become simpler because the business domain object you pulled out is now a dependency that you _can_ stub out.

**Rule of thumb**:

* When extracting behaviour, try to end up with a more business centric object, and a more generic
  object that accepts the new dependency. **TBD**: example here


# Drawbacks
[drawbacks]: #drawbacks
TBD:
Why should we *not* do this?

# Alternatives
[alternatives]: #alternatives
TBD
- What other designs have been considered?
- Why is this proposal the best?
- What is the impact of not doing this?

# Prior Art
[prior-art]: #prior-art

TBD: plenty more examples out there.

## Rails
Ruby on Rails uses onion skin tests, which work relatively well except that in Rails, the
innermost layer is async and non-deterministic. This can lead to huge surprises when unit tests
assume friendly actors upon the database, especially when table/row locking is insufficient for
concurrency safety.

## Hexagonal Architecture
TBD: An example where a repository style pattern is used to lift IO to the controller layer.

# Unresolved Questions
[unresolved-questions]: #unresolved-questions

## Feature tests
* can we provide an RFC that describes feature tests which test typical acceptance criteria, so that we
  can trust those an minimize 'over-excercising' our units.
