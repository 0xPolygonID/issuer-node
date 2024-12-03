export type DisplayMethod = {
  id: string;
  name: string;
  url: string;
};

export type DisplayMethodMetadata = {
  backgroundImageUrl: string;
  description: string;
  descriptionTextColor: string;
  issuerName: string;
  issuerTextColor: string;
  logo: {
    alt: string;
    uri: string;
  };
  title: string;
  titleTextColor: string;
};
