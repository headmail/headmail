declare module 'mjml-browser' {
  const mjml2html: (mjml: string, opts?: any) => { html: string; errors?: any[] };
  export default mjml2html;
}
