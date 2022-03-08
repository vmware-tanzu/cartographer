---
title: "A live editor for Supply Chains"
slug: live-editor
date: 2022-02-10
author: Rasheed Abdul-Aziz
authorLocation: https://github.com/squeedee
image: /img/posts/live-editor/cover-image.png
excerpt: "A scratchpad for editing Supply Chains"
tags: ["Live editor", "Supply Chains"]
---

# >> [Click to Launch editor](/live-editor/index.html) <<

## What is this?

I recently decided I wanted JSON Schemas for all of Cartographer's CRDs. That way I'd be less likely to make mistakes
and need to wait until I had done a `kubectl apply` and waited for a **workload** failure to discover that I had
forgotten something important.

I wrote a little bit of code to generate JSON Schemas from
[Cartographer CRD manifests](https://github.com/vmware-tanzu/cartographer/tree/main/config/crd/bases). I plan on
releasing it as a formal utility that folks can use to create JSON Schemas for any OpenAPIV3 CRD Schemas, but I realised
**schema validation is not enough**.

So I've created the [Live Editor](/live-editor/index.html).

Today the editor:

- Only supports SupplyChains.
- Visualises the supply chain.
- AutoSuggests based on schema.
- Validates based on schema.
- AutoSuggests inputs from existing resources.
- Let's you copy the URL to share your yaml with others.

In the future I'd like to add:

- Support for all the CRDs
- Syntax checking for invalid references (coming soon!)
- Display `options` selected by a workload (part of a new set of RFCs to support multiple paths in a supply chain)
- Instances of the editor inline in the docs
- Links in example CRs to editable/visualized versions of the same CR
- Read/Write GitHub Gists to support loading and editing templates + blueprints + owners
- Much more!

I want to know what you think! Please let us know if this is something that's valuable to you.

Please provide feedback by
[commenting or upvoting in this issue](https://github.com/vmware-tanzu/cartographer/issues/566)!
