import { Flex, Image, Typography } from "antd";
import { useEffect, useRef, useState } from "react";
import { processUrl } from "src/adapters/api/schemas";

import IconMore from "src/assets/icons/more.svg?react";
import { useEnvContext } from "src/contexts/Env";
import { DisplayMethodMetadata } from "src/domain";

export const DISPLAY_METHOD_CARD_WIDTH = 500;
export const DIMENSIONS_TO_FONT_RATIO = 20;
const DEFAULT_FONT_SIZE = 18;

function getEmSize(size: number) {
  return `${size / DEFAULT_FONT_SIZE}em`;
}

export function DisplayMethodCard({ metadata }: { metadata: DisplayMethodMetadata }) {
  const elementRef = useRef(null);
  const env = useEnvContext();
  const [dimensions, setDimensions] = useState({ left: 0, right: 0 });

  const fontSize = (dimensions.left + dimensions.right) / DIMENSIONS_TO_FONT_RATIO;

  const processedBackgroundImageUrl = processUrl(metadata.backgroundImageUrl, env);
  const processedLogoImageUrl = processUrl(metadata.logo.uri, env);

  useEffect(() => {
    const resizeObserver = new ResizeObserver((entries) => {
      for (const entry of entries) {
        const { left, right } = entry.contentRect;
        setDimensions({ left, right });
      }
    });

    if (elementRef.current) {
      resizeObserver.observe(elementRef.current);
    }

    return () => resizeObserver.disconnect();
  }, []);

  return (
    <Flex
      className="background-grey"
      ref={elementRef}
      style={{
        aspectRatio: 1.586,
        backgroundImage: `url(${processedBackgroundImageUrl.success ? processedBackgroundImageUrl.data : metadata.backgroundImageUrl})`,
        backgroundSize: "cover",
        borderRadius: getEmSize(14),
        fontSize,
        maxWidth: DISPLAY_METHOD_CARD_WIDTH,
        padding: `${getEmSize(18)} ${getEmSize(20)} ${getEmSize(24)}`,
        width: "100%",
      }}
      vertical
    >
      <Typography.Title
        style={{
          color: metadata.titleTextColor,
          fontSize: getEmSize(18),
          fontWeight: 600,
          margin: 0,
        }}
      >
        {metadata.title}
      </Typography.Title>
      <Typography.Text
        style={{
          color: metadata.descriptionTextColor,
          fontSize: getEmSize(14),
          marginTop: getEmSize(8),
        }}
      >
        {metadata.description}
      </Typography.Text>

      <Flex align="center" gap={getEmSize(8)} style={{ marginTop: "auto" }}>
        <Image
          alt={metadata.logo.alt}
          preview={false}
          src={processedLogoImageUrl.success ? processedLogoImageUrl.data : metadata.logo.uri}
          style={{ fontSize, width: getEmSize(44) }}
        />
        <Flex vertical>
          <Typography.Text
            style={{ color: metadata.issuerTextColor, fontSize: getEmSize(14), opacity: 0.7 }}
          >
            Issuer
          </Typography.Text>
          <Typography.Text
            style={{ color: metadata.issuerTextColor, fontSize: getEmSize(16), fontWeight: 600 }}
          >
            {metadata.issuerName}
          </Typography.Text>
        </Flex>
        <IconMore
          color={metadata.issuerTextColor}
          style={{ marginLeft: "auto" }}
          width={getEmSize(24)}
        />
      </Flex>
    </Flex>
  );
}
