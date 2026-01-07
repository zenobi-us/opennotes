import type { ArkErrors } from 'arktype';

/**
 * Pretty print Arktype errors in a hierarchical format grouped by path.
 *
 * Example output:
 * - user.email
 *   - must be an email address (was "test")
 * - user.age
 *   - must be more than 13 (was 12)
 */
export const prettifyArktypeErrors = (errors: ArkErrors): string => {
  const grouped = new Map<string, string[]>();

  for (const error of errors) {
    const path = error.path?.join('.') || '(root)';
    const message = error.toString();

    if (!grouped.has(path)) {
      grouped.set(path, []);
    }
    grouped.get(path)!.push(message);
  }

  const lines: string[] = [];

  for (const [path, messages] of grouped) {
    lines.push(`- ${path}`);
    for (const message of messages) {
      lines.push(`  - ${message}`);
    }
  }

  return lines.join('\n');
};
