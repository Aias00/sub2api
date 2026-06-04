# Model Plaza

Model Plaza is the public model catalog page, typically exposed at `/models`.

Its role is not to replace the console, but to help you quickly judge:

- which model families the site currently highlights
- what each model is good at
- which capability or pricing notes are publicly visible

## How Model Plaza differs from groups

Keep this distinction clear:

- Model Plaza is the **public presentation layer**
- groups are the **actual authorization layer**

So a model being visible in Model Plaza does not automatically mean your current key can use it.

Actual access still depends on:

- your available groups
- the group bound to your key
- the current site policy

## Good uses for Model Plaza

### Public preview before login

If you only want to know whether the site is a fit for your workflow, Model Plaza is the fastest entry point.

### Team alignment

It gives one shared answer to questions like:

- which model do we recommend for daily coding?
- which one is better for heavy reasoning?
- which one is more specialized?

### Pre-purchase screening

Before buying anything or creating a key, you can use Model Plaza to decide whether the available model mix is right for you.

## When not to rely on Model Plaza alone

If you are already signed in and about to make real requests:

1. check [Available Groups](#/console/available-groups)
2. create or inspect [API Keys](#/console/api-keys)
3. use the console’s API Guide for the final configuration

That is the full path from public information to actual access.

## Where the content comes from

Model Plaza is driven by backend configuration rather than hard-coded frontend content.  
That means titles, tags, descriptions, and price notes may change as the site operator updates the catalog.
