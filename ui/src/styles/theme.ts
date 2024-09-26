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
  primaryColorDark: string;
  primaryColorLight: string;
  successBg: string;
  successColor: string;
  successLightColor: string;
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
    primaryColorDark: z.string(),
    primaryColorLight: z.string(),
    successBg: z.string(),
    successColor: z.string(),
    successLightColor: z.string(),
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
  bgLight,
  borderColor,
  errorColor,
  primaryColor,
  primaryColorDark,
  primaryColorLight,
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
      colorPrimary: primaryColorLight,
      colorPrimaryBorder: primaryColor,
      colorPrimaryText: textColorSecondary,
      controlHeight: 40,
      defaultBorderColor: primaryColorDark,
      defaultColor: primaryColorDark,
      paddingContentHorizontal: 16,
      primaryColor: textColor,
      primaryShadow: "none",
    },
    Card: {
      colorBgBase: primaryColor,
      fontWeightStrong: 500,
    },
    DatePicker: { controlHeight: 40 },
    Form: { fontSize: 14 },
    Input: { controlHeight: 40 },
    InputNumber: { controlHeight: 40 },
    Layout: { bodyBg: "white", headerBg: "white", siderBg: "white" },
    Menu: {
      colorBgBase: "transparent",
      horizontalItemSelectedColor: primaryColor,
      itemActiveBg: "transparent",
      itemBg: "transparent",
      itemColor: textColorSecondary,
      itemHoverBg: "white",
      itemHoverColor: primaryColor,
      itemSelectedBg: successBg,
      itemSelectedColor: primaryColor,
      subMenuItemBg: "white",
    },
    Message: { fontSize: 18 },
    Radio: {
      buttonCheckedBg: bgLight,
      colorPrimary: primaryColor,
      colorPrimaryHover: primaryColor,
      controlHeight: 40,
      size: 20,
      wrapperMarginInlineEnd: 0,
    },
    Select: { colorBorder: primaryColor, controlHeight: 40 },
    Table: { fontSize: 14, fontWeightStrong: 400 },
    Tabs: {
      colorPrimary: primaryColor,
      itemHoverColor: primaryColor,
    },
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
    colorLink: primaryColor,
    colorLinkActive: primaryColor,
    colorLinkHover: primaryColor,
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
