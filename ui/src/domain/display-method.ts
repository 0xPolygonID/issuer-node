export enum DisplayMethodType {
  Iden3BasicDisplayMethodV1 = "Iden3BasicDisplayMethodV1",
}

export type DisplayMethod = {
  id: string;
  name: string;
  type: DisplayMethodType;
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
