/**
 * Slugify text: convert to lowercase, remove special chars, replace spaces with hyphens
 */
export function slugify(text: string): string {
  return text
    .toLowerCase()
    .replace(/\n/g, ' ')
    .replace(/[^a-z0-9\s-]/g, '')
    .replace(/\s+/g, '-')
    .replace(/^-+|-+$/g, '');
}

/**
 * Remove leading whitespace from multiline strings
 */
export function dedent(text: string): string {
  const lines = text.split('\n');
  const indent = lines.find((line) => line.trim())?.match(/^\s*/)?.[0]?.length ?? 0;
  return lines.map((line) => line.slice(indent)).join('\n');
}

/**
 * Render a string with template variables
 */
export function renderTemplateString(
  template: string,
  variables: Record<string, string | number | boolean>
): string {
  return template.replace(/\{\{(\w+)\}\}/g, (_, key) => {
    const value = variables[key];
    return value !== undefined ? String(value) : '';
  });
}

export function objectToFrontmatter(obj: Record<string, any>): string {
  const entries = Object.entries(obj).map(([key, value]) => {
    if (Array.isArray(value)) {
      return `${key}:\n${value.map((item) => `  - ${item}`).join('\n')}`;
    } else {
      return `${key}: ${value}`;
    }
  });
  return entries.join('\n');
}
