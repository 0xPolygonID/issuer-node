import { ThemeConfig } from "antd/es/config-provider/context";
import { z } from "zod";

import { getStrictParser } from "src/adapters/parsers";
import variables from "src/styles/variables.module.scss";

type StyleVariables = {
  avatarBg: string;
  bgLight: string;
  borderColor: string;
  dividerColor: string;
  errorBg: string;
  errorColor: string;
  iconBg: string;
  iconColor: string;
  primaryBg: string;
  primaryColor: string;
  successBg: string;
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
    dividerColor: z.string(),
    errorBg: z.string(),
    errorColor: z.string(),
    iconBg: z.string(),
    iconColor: z.string(),
    primaryBg: z.string(),
    primaryColor: z.string(),
    successBg: z.string(),
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
  borderColor,
  errorColor,
  primaryColor,
  successBg,
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
    Button: {
      colorBgContainerDisabled: successBg,
      colorPrimaryBg: primaryColor,
      colorPrimaryHover: "#74F526",
      controlHeight: 40,
      defaultHoverBorderColor: successColor,
      defaultHoverColor: successColor,
      paddingContentHorizontal: 16,
      primaryColor: textColor,
    },
    Card: {
      colorBgBase: successColor,
      fontWeightStrong: 500,
    },
    Checkbox: { borderRadius: 6, colorPrimary: successBg, colorWhite: "#3AB000", size: 20 },
    DatePicker: { controlHeight: 40 },
    Form: { fontSize: 14 },
    Input: { controlHeight: 40 },
    InputNumber: { controlHeight: 40 },
    Layout: { bodyBg: "white", headerBg: "white", siderBg: "white" },
    Menu: {
      colorBgBase: "transparent",
      horizontalItemSelectedColor: successColor,
      itemActiveBg: "transparent",
      itemBg: "transparent",
      itemColor: textColorSecondary,
      itemHoverBg: "white",
      itemHoverColor: successColor,
      itemSelectedBg: successBg,
      itemSelectedColor: successColor,
      subMenuItemBg: "white",
    },
    Message: { fontSize: 18 },
    Radio: { controlHeight: 40, size: 20 },
    Select: { colorBorder: successColor, controlHeight: 40 },
    Table: { fontSize: 14, fontWeightStrong: 400 },
    Tag: {
      colorBgBase: tagBg,
      colorTextBase: tagColor,
    },
    Typography: {
      colorWarning: textColorWarning,
      fontSize: 14,
    },
  },
  token: {
    borderRadius: 8,
    colorError: errorColor,
    colorInfo: tagColor,
    colorInfoBorder: borderColor,
    colorLink: successColor,
    colorLinkActive: successColor,
    colorLinkHover: successColor,
    colorPrimary: primaryColor,
    colorPrimaryText: textColor,
    colorSuccess: successColor,
    colorText: textColor,
    colorTextLabel: tagColor,
    colorTextSecondary: textColorSecondary,
    fontFamily: "Matter-Regular",
    fontSize: 16,
    fontSizeHeading2: 32,
    fontSizeSM: 14,
  },
};
