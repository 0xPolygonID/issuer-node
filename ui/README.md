# Self-hosted Polygon ID Issuer Node UI

The frontend application entirely dependent on the [Self-hosted Polygon ID Issuer Node](../README.md).

## Installation

Configure, install and set up the [issuer node](../README.md#installation) first and follow the instructions for the UI.

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
  - `format` (any string)
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

## License

See [LICENSE](../LICENSE.md).
