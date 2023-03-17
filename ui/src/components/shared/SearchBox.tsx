import { Input } from "antd";
import { useEffect, useState } from "react";

import { ReactComponent as IconSearch } from "src/assets/icons/search-lg.svg";
import { Spinner } from "src/components/shared/Spinner";
import { DEBOUNCE_INPUT_TIMEOUT } from "src/utils/constants";

export function SearchBox({
  onSearch,
  placeholder,
  query,
}: {
  onSearch: (value: string) => void;
  placeholder: string;
  query: string | null;
}) {
  const [isSearching, setIsSearching] = useState<boolean>(false);
  const [searchValue, setSearchValue] = useState<string | null>(query);

  useEffect(() => {
    if (searchValue === null || searchValue === query) {
      return;
    }

    setIsSearching(true);

    const debounceSearch = setTimeout(() => {
      onSearch(searchValue);
      setIsSearching(false);
    }, DEBOUNCE_INPUT_TIMEOUT);

    return () => {
      clearTimeout(debounceSearch);
      setIsSearching(false);
    };
  }, [onSearch, query, searchValue]);

  useEffect(() => {
    setSearchValue((oldSearchValue) => {
      if (oldSearchValue !== query) {
        return query;
      } else {
        return oldSearchValue;
      }
    });
  }, [query]);

  return (
    <Input
      onChange={({ target: { value } }) => setSearchValue(value)}
      placeholder={placeholder}
      prefix={isSearching ? <Spinner size="default" /> : <IconSearch />}
      value={searchValue || undefined}
    />
  );
}
