import { ThemeConfig } from "antd/es/config-provider/context";
import { z } from "zod";

import { getStrictParser } from "src/adapters/parsers";
import variables from "src/styles/variables.module.scss";

type StyleVariables = {
  avatarBg: string;
  bgLight: string;
  borderColor: string;
  cyanBg: string;
  cyanColor: string;
  dividerColor: string;
  errorBg: string;
  errorColor: string;
  primaryBg: string;
  primaryColor: string;
  successColor: string;
  tagBg: string;
  tagBgSuccess: string;
  tagColor: string;
  textColor: string;
  textColorSecondary: string;
  textColorWarning: string;
};

const parsedStyleVariables = getStrictParser<StyleVariables>()(
  z.object({
    avatarBg: z.string(),
    bgLight: z.string(),
    borderColor: z.string(),
    cyanBg: z.string(),
    cyanColor: z.string(),
    dividerColor: z.string(),
    errorBg: z.string(),
    errorColor: z.string(),
    primaryBg: z.string(),
    primaryColor: z.string(),
    successColor: z.string(),
    tagBg: z.string(),
    tagBgSuccess: z.string(),
    tagColor: z.string(),
    textColor: z.string(),
    textColorSecondary: z.string(),
    textColorWarning: z.string(),
  })
).parse(variables);

const {
  avatarBg,
  errorColor,
  primaryColor,
  successColor,
  tagBg,
  tagColor,
  textColor,
  textColorSecondary,
  textColorWarning,
} = parsedStyleVariables;

export const theme: ThemeConfig = {
  components: {
    Avatar: { colorBgBase: avatarBg },
    Button: { controlHeight: 40, paddingContentHorizontal: 16 },
    Card: { fontWeightStrong: 500 },
    Checkbox: { borderRadius: 6, size: 20 },
    DatePicker: { controlHeight: 40 },
    Form: { fontSize: 14 },
    Input: { controlHeight: 40 },
    InputNumber: { controlHeight: 40 },
    Layout: { bodyBg: "white", headerBg: "white", siderBg: "white" },
    Menu: {
      itemHoverBg: "white",
      itemHoverColor: primaryColor,
      subMenuItemBg: "white",
    },
    Message: { fontSize: 18 },
    Radio: { controlHeight: 40, size: 20 },
    Select: { controlHeight: 40 },
    Table: { fontSize: 14, fontWeightStrong: 400 },
    Tag: {
      colorBgBase: tagBg,
      colorTextBase: tagColor,
    },
    Typography: {
      colorWarning: textColorWarning,
    },
  },
  token: {
    borderRadius: 8,
    colorError: errorColor,
    colorInfo: primaryColor,
    colorLink: primaryColor,
    colorLinkActive: primaryColor,
    colorLinkHover: primaryColor,
    colorPrimary: primaryColor,
    colorSuccess: successColor,
    colorText: textColor,
    colorTextLabel: tagColor,
    colorTextSecondary: textColorSecondary,
    fontFamily: "ModernEra-Regular",
    fontSize: 16,
    fontSizeHeading2: 32,
    fontSizeSM: 14,
  },
};
