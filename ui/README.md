# Privado ID Issuer Node UI

The frontend application of the [Privado ID Issuer Node](../README.md).

## Installation

1. Configure, install and set up the [issuer node](../README.md#installation), following the optional steps about setting up the UI
2. _Optional_ Follow the instructions for [developing the UI](../README.md#development-ui).

## Built with

- [Ant](https://ant.design)
- [React](https://reactjs.org)
- [Typescript](https://www.typescriptlang.org)
- [Vite](https://vitejs.dev)
- [Docker](https://docs.docker.com/get-started/)

## JSON Schema support

This application supports a subset of the features of the [JSON Schema](https://json-schema.org/) standard draft 2020-12. Older drafts will normally work, as long as they don't make use of colliding features that are incompatible with the draft 2020-12.

All basic types are supported along with `title` and `description` keywords. The following keywords are also supported by type:

- `string`
  - `enum` (partially)
  - `format`
    - You can use any of the standard formats (`date-time`, `time`, `date`, `duration`, `email`, `idn-email`, `hostname`, `idn-hostname`, `ipv4`, `ipv6`, `uuid`, `uri`, `uri-reference`, `iri`, `iri-reference`, `uri-template`, `json-pointer`, `relative-json-pointer`, `regex`) or any other custom `string`, but only `date-time`, `time` and `date` will show specialized inputs.
- `number`
  - `enum` (partially)
- `integer`
  - `enum` (partially)
- `boolean`
  - `enum` (partially)
- `object`
  - `properties`
  - `required`
- `array`
  - `items`
- `null`

As described above, the `enum` keyword is partially supported to fit use cases that make sense in this application, i.e basically a dropdown list with limited choices. The support of the `enum` is as follows:

- Only `string`, `number`, `integer` and `boolean` schemas support `enum`.
- Only values that validate against the schema are allowed.
- `enum` only schemas (without a type) will not pass validation.
- Repeated values won't produce an error.

While the UI correctly parses all the schemas above, currently there's a limitation regarding the issuance of credentials since the issue credential form does not yet support the schemas `array` and `null`. This, in practice, means that while you can import schemas that declare attributes of type `array` and `null`, such attributes will not be operative in the issue credential form, rendering the credential not-issuable in practice when the attributes are required, since the user will not be able to provide a value for them in the UI.

## License

See [LICENSE](../LICENSE.md).
