# rfc

[Cartographer] Requests for Comments.

# Motivation
* To separate design discussions from daily feature work.
* To ensure the backlog only contains ratified work to be done.
* To support healthy dialog, by: 
  * supporting asynchronous discussion
  * tracking design decisions and motivations
  * sharing design decisions and motivations with all contributors and stakeholders

# Process
(Once we're OSS on Github, we will automate as much of this as possible)

To start an RFC:
## Claim an RFC ID
* Checkout `main`
* Add the next sequential RFC number, and a working title, as a file in the form `rfc-<0000>-<working-title>.md`
* Add, commit, and push (with commit message `creating RFC <0000> working title`)
  * If you hit a commit error, rebase and ensure your RFC ID doesn't clash.

## Start a Draft
* Check out a branch named the same as your file (except for `.md`): `rfc-<0000>-<working-title>`
* Fill in your draft RFC. 
  * Describe the **Motivation** for the RFC
  * Try to provide some **possible solutions** and any **cross references** to help everyone better understand the domain space. 
  * You can start off with a copy of [draft-template.md].  
* Make a merge request
  * Commit with message `Draft: introducing RFC <0000> working title`
  * Push your branch upstream
  * Make an MR and ensure the title begins with `draft:` (so we know not to merge it).
* Update the draft document on `main` to contain a link to the RFC
  ```markdown
  [Draft MR](https://github.com/vmware-tanzu/cartographer/-/merge_requests/<mr-number>)
  ```


## Maintain the Draft

* Let folks know about it! Link them to the MR where they can discuss it! 
* Try and merge valuable suggestions into the body of the RFC.


## Find a good solution

You're driving to a final proposed solution, so as soon as you're comfortable:

1. rewrite the RFC to match [final-template.md]
1. remove the `draft: ` label from the MR
2. ask that the core team reviews and merges the RFC.

There may still be a considerable amount of discussion after you finalize the PR.
The core team might merge it and still not plan to implement it, but they will communicate
where it sits on the roadmap.

## Core team schedule

The core team has two-weekly meetings to monitor the state of RFCs and to help ensure that contributors 
receive feedback or at least status updates. 

[final-template.md]: ./final-template.md  "Final Template"
[draft-template.md]: ./draft-template.md  "Draft Template"
[Cartographer]: https://cartographer.sh  "Cartographer home page"
