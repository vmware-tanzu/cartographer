# Meta
[meta]: #meta
- Name: RFC Process
- Start Date: 2022-01-24
- Author(s): @martyspiewak
- Status: Draft <!-- Acceptable values: Draft, Approved, On Hold, Superseded -->
- RFC Pull Request: (leave blank)
- Supersedes: N/A

# Summary
[summary]: #summary

The "RFC" (request for comments) process provides a consistent and controlled path for new features to enter the project so that all stakeholders can be confident about the direction Cartographer is evolving in.

# Motivation
[motivation]: #motivation

Now that we've started releasing minor versions and people are trying out Cartographer, we need to take more care in introducing change. This process documents a path for anyone to contribute and have input on features.

# What it is
[what-it-is]: #what-it-is

RFCs record potential change with the context and information at the given time. This provides a defined process for anyone wishing to contribute to the Cartographer project and gives opportunity for engagement. Anyone who chooses not to actively partcipate in any RFC is presumed to trust their colleagues on the matter. Once an RFC is accepted, they can be referenced as read-only documents in this repository until replaced or ammended by another RFC when context has significantly changed.

# How it Works
[how-it-works]: #how-it-works

Many changes, including bug fixes and documentation improvements can be implemented and reviewed via the normal GitHub pull request workflow.

Some changes though are "substantial", and we ask that these be put through a bit of a design process and produce a consensus among the community and the maintainer team.

## What's in Scope

You'll need to follow this process for anything considered "substantial". The things that we consider "substantial" will evolve over time, but will generally be user facing changes.

What constitutes a "substantial" change may include the following but is not limited to:

- changes to a resource's spec
- changes to a resource's status
- breaking changes
- governance changes
- issues that need more discussion or clarification as determined by the project maintainers

If you submit a pull request or issue that the team deems warrants an RFC, it will be politely closed with a request for an RFC.

## Process

### RFCs

To get an RFC into Cartographer, first the RFC needs to be merged into the repo. Once an RFC is merged, it's considered 'active' and may be implemented to be included in the project. These steps will get an RFC to be considered:

- Fork the repo: <https://github.com/vmware-tanzu/cartographer>
- Copy `rfc/rfc-0000-template.md` to `rfc/rfc-0000-my-feature.md` (where 'my-feature' is descriptive. don't assign an RFC number yet).
- Fill in RFC. Any section can be marked as "N/A" if not applicable.
- Submit a pull request. The pull request is the time to get review of the proposal from the larger community.
- Build consensus and integrate feedback. RFCs that have broad support are much more likely to make progress than those that don't receive any comments.

Once a pull request is opened, the RFC is now in development and the following will happen:

- It will be discussed in a future office hours meeting. Office hours happen on a weekly cadence barring exceptions.
- The team will discuss as much as possible in the RFC pull request directly. Any outside discussion will be summarized in the comment thread.
- When deemed "ready", a team member will propose a "motion for final comment period (FCP)" along with a disposition of the outcome (merge, close, or postpone). This is step taken when enough discussion of the tradeoffs have taken place and the team is in a position to make a decision. Before entering FCP, a super majority of the team must sign off.
- The FCP will last 7 days. If there's unanimous agreement among the team the FCP can close early.
- For voting, the binding votes are comprised of the maintainer team and Technical Oversight Committee. Acceptance requires a super majority of maintainer votes and at least one Technical Oversight Committee vote in favor. The voting options are the following: Affirmative, Negative, and Abstain. Non-binding votes are of course welcome. Super majority means 2/3 or greater of binding votes. (Abstentions are not counted in the calculation)
- If no substantial new arguments or ideas are raised, the FCP will follow the outcome decided. If there are substantial new arguments, then the RFC will go back into development.

Once an RFC has been accepted, the team member who merges the pull request should do the following:

- Assign an id based off the latest available number.
- Rename the file based off the id inside `rfc/`.
- Fill in the remaining metadata at the top.
- Commit everything.

### Status Field
The status field in the meta section has the following acceptable values:
- **Draft**: The maintainer team has not yet approved the RFC. Edits are still being made.
- **Approved**: The maintainer team has approved the RFC and is accepting PRs for its implementation.
- **On Hold**: The maintainer team has approved the RFC, but PRs are currently not being accepted for its implementation. A reason must be provided in the RFC.
- **Superseded**: The maintainer team has approved the RFC, but it is replaced by another RFC.

RFCs that are rejected will no longer be pulled into the `rfc/` directory.

# Drawbacks
[drawbacks]: #drawbacks

Though RFCs are intended to be lightweight, this introduces extra process than what we're doing today. The additional work may make consistency hard.
In not archiving rejected/unapproved RFCs, Cartographer may lose context of past discussions and decisions.

# Alternatives
[alternatives]: #alternatives

- Stick with the current process, which is not well-defined, does not include a formal voting on approval of RFCs, and does not clearly indicate whether RFCs are in draft, accepted, rejected, or implemented. This makes it hard for new people to know how to propose changes, and how things are decided. It also makes it hard to track major changes to the project.

# Prior Art
[prior-art]: #prior-art

The basic format of RFCs was [invented in the 1960s by Steve Crocker](https://en.wikipedia.org/wiki/Request_for_Comments#History). The current RFC process has been derived heavily from [Buildpacks](https://github.com/buildpacks/rfcs/). Many other open source projects have adopted an RFC process:

- [Rust](https://github.com/rust-lang/rfcs)
    - FCP lasts 10 days
- [Ember](https://github.com/emberjs/rfcs)
    - FCP lasts 7 days
- [Yarn](https://github.com/yarnpkg/rfcs)
    - FCP lasts 7 days
- [React](https://github.com/reactjs/rfcs)
    - FCP lasts 3 days
- [IETF](https://www.rfc-editor.org/rfc/rfc2026.txt)

