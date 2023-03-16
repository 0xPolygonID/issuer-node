export function kebabCase(string: string) {
  const kebab = string.match(
    /[A-Z]{2,}(?=[A-Z][a-z]+[0-9]*|\b)|[A-Z]?[a-z]+[0-9]*|[A-Z]|[0-9]+/g
  ) || ["-"];

  return kebab.join("-").toLowerCase();
}
