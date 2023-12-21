# ![MACH composer](./_img/logo.png)

This documentation describes the workings of MACH composer. Intended to set up
and manage a **M**icroservices based, **A**PI-first, **C**loud-native SaaS and
**H**eadless platform.

{%
include-markdown "overhaul_warning.md"
%}

## What is it?

MACH composer is a framework that you use to orchestrate and extend modern
digital commerce & experience platforms, based on MACH technologies and cloud
native services. It provides a standards-based, future-proof tool-set and
methodology to hand to your teams when building these types of platforms.

It includes:

- A configuration framework for managing MACH-services configuration, using
  infrastructure-as-code underneath (powered by Terraform)
- A microservices architecture based on modern serverless technology (AWS Lambda
  and Azure Functions), including (alpha) support for building your
  microservices
  with the Serverless Framework
- Multi-tenancy support for managing many instances of your platform, that share
  the same library of microservices
- CI/CD tools for automating the delivery of your MACH ecosystem
- Tight integration with AWS, Azure and GCP
- Opt-in functionality for smart management of dependencies between components,
  reducing deployment time and risk

The framework is intended as the 'centerpiece' of your MACH architecture and
incorporates industry best practises such as the 12 Factor Methodology,
Infrastructure-as-code, DevOps, immutable deployments, FAAS, etc.

With combining (and requiring) these practises, using the framework has
significant impact on your engineering methodology and organisation. On the
other hand, by combining those practises we believe it offers an accelerated
'way in' in terms of embracing modern engineering practises in your
organisation.

## Documentation structure

- [Tutorials](./tutorial/index.md) introduces you to MACH composer and lets you
  set up your MACH stack in a couple of steps
- [Explanations](concepts/index.md) explanations of concepts, best practises
  and techniques
- [Reference guides](./reference/index.md) contain technical reference for the
  MACH syntax and usage of the CLI.
- [How-to guides](./howto/index.md) contain practical descriptions on how to
  solve certain problems. They are more advanced than tutorials and assume some
  knowledge of how MACH works.

## Where from here?

- Start by setting up your [first MACH stack](./tutorial/index.md)
- Read more about the [MACH composer concepts](concepts/index.md)
- Wonder how to reason about your MACH stack? Read
  our[Architectural Guidance](concepts/architecture/index.md)
