---
title: "Why Cartographer? A pre-history"
slug: why-cartographer
date: 2022-02-01
author: Rasheed Abdul-Aziz
authorLocation: https://github.com/squeedee
image: /img/posts/why-cartographer/history-console.png
excerpt: "How we got here, Heroku, Cloud Foundry and the 12 Factor App"
tags: ['Cloud Foundry', 'Heroku', '12 Factor App']
---

## 12 Factor Apps and Heroku

Over a decade ago, a seismic shift occurred in the history of online application development. Heroku
introduced `buildpacks`, the `git push heroku` idiom and most importantly, the [12 factor app](https://12factor.net/).

In this brave new world, developers write applications that adhere to straight-forward constraints (the 12 factors) to
ensure the apps can be deployed almost anywhere. With a simple `git push` to a different origin, the app is transformed
by the buildpacks, having their base dependencies injected into a disk image that can then be deployed and scaled out on
virtual machines (and more recently containers). Furthermore, the app's can easily define their service dependencies (
such as databases) and be dynamically connected to the correct service instance for the environment (such as dev,
staging and production).

This contrasts with developers spending significant time developing deployment scripts to ensure their app will run on
some (often bespoke) infrastructure.

The full benefits of Heroku's 12 Factor model are too numerous to delve into here, so here is a [list of resources](tbd)
you can dive into at your leisure.

There were tradeoffs with Heroku's model, such as a certain infelxibility to how you design your app and what could be
run as a sidecar (vs) a service. However, it smoothed the path to production for startups so greatly, that a lot of
enterprise staff, engineers and operators alike, wanted the same "magic" in the enterprise.

This was a significant limitation of Heroku. It was publicly hosted, did not support "air gap environments", and was
private IP.

## Cloud Foundry

Cloud foundry is one of the efforts (originally started at VMware) to bring the Heroku 12 Factor App into the
enterprise. It differed from Heroku by:

1. Supporting Multiple IaaS's and generalized internal infrastructure (such as vsphere)
2. Open Source first
3. Buildpacks that support `offline mode` - dependency injection into the image without reaching out to the internet.
4. A deployment orchestration system, including an operator dashboard.

Cloud Foundry let developers use the Heroku idiom with ease, and let operators maintain control of security and
infrastructure, without the arcane 'over the wall' mentality that was hindering modern app development in the
enterprise.

A couple of concerns about Cloud foundry that Cartographer want's to adress:

1. BOSH as a Infrastructure adaptor, although open source, has not seen significant adoption. Kubernete's on the other
   hand has.
2. Cloud Foundry `cf push` is highly prescriptive and unchangeable. Companies need a mechanism to modify parts of the
   process to fulfill operational and migration needs.

## Cartographer

Perhaps you've seen a Cloud Foundry manifest:


