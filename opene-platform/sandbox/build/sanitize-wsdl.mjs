// Replace endpoint defaults and remote WSDL/schema imports in packaged vendor
// contracts. Namespaces and locally bundled imports remain untouched. A real
// simulated endpoint must still be configured before a provider is enabled.
export function sanitizeWsdl(input) {
  if (input.includes('\0')) throw Error('Unsupported non-ASCII-compatible contract encoding');
  let changed = 0;
  const replaceUrl = (tag) => tag.replace(/(\b(?:location|schemaLocation)\s*=\s*["'])(https?:\/\/[^"']+)(["'])/gi,
    (all, before, url, after) => {
      if (url === 'http://127.0.0.1:1/disabled') return all;
      changed++;
      return before + 'http://127.0.0.1:1/disabled' + after;
    });
  const result = input.replace(/<(?:[\w.-]+:)?(?:address|import|include|redefine)\b[^>]*>/gi, replaceUrl);
  return { result, changed };
}
