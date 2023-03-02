# Self-hosted Polygon ID Platform UI

The frontend application of [Self-hosted Polygon ID Platform](../README.md).

## Installation

1. Copy `.env.example` as `.env`, and amend the endpoint variable as needed.
1. Run `npm install` to install.
1. Run `npm run prepare` to set up commit hooks.
1. Run `npm start` to build & run locally.

The application will be available on [http://localhost:5173](http://localhost:5173).

## Built with

- [Ant](https://ant.design)
- [React](https://reactjs.org)
- [Typescript](https://www.typescriptlang.org)
- [Vite](https://vitejs.dev)
- [Docker](https://docs.docker.com/get-started/)

## Docker image

To locally generate a Docker image of the Polygon ID Platform, you can run the following command:

```sh
docker build . -t platform:local
```

The Docker image won't build the application until you run it, to allow dynamic environment variables to be passed via the command. This facilitates the deployment process. The environment variables that you need to pass to the
`docker run` command are the same as those in the `.env.example` file but without the `VITE_` prefix.

Example:

```sh
docker run \
-e API=https://api-staging.polygonid.com/v1 \
-p 8080:80 platform:local
```

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
