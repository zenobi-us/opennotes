import { marked } from 'marked';
import TuiRenderer from 'marked-terminal';
import { renderTemplateString } from '../core/strings.ts';

marked.setOptions({
  renderer: new TuiRenderer(),
});

const RenderMarkdownTui = async (
  markdown: string,
  variables?: Record<string, string | number | boolean>
): Promise<string> => {
  const rendered = variables ? renderTemplateString(markdown, variables) : markdown;
  return await marked(rendered);
};

export { RenderMarkdownTui };
