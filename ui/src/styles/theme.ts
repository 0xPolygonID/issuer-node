import { ThemeConfig } from "antd/es/config-provider/context";
import z from "zod";

import { StyleVariables } from "src/domain";
import variables from "src/styles/variables.module.scss";
import { StrictSchema } from "src/utils/types";

export const parseStyleVariables = StrictSchema<StyleVariables>()(
  z.object({
    avatarBg: z.string(),
    bgLight: z.string(),
    borderColor: z.string(),
    cyanBg: z.string(),
    cyanColor: z.string(),
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
  })
);

const parsedStyleVariables = parseStyleVariables.safeParse(variables);

if (!parsedStyleVariables.success) {
  throw new Error("Invalid style variables");
}

const {
  avatarBg,
  errorColor,
  primaryColor,
  successColor,
  tagBg,
  tagColor,
  textColor,
  textColorSecondary,
} = parsedStyleVariables.data;

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
    Layout: { colorBgBody: "white", colorBgHeader: "white" },
    Menu: {
      colorItemBgHover: "white",
      colorItemTextHover: primaryColor,
      colorSubItemBg: "white",
    },
    Message: { fontSize: 18 },
    Radio: { controlHeight: 40, size: 20 },
    Select: { controlHeight: 40 },
    Table: { fontSize: 14, fontWeightStrong: 400 },
    Tag: {
      colorBgBase: tagBg,
      colorTextBase: tagColor,
      fontSize: 14,
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
  },
};
