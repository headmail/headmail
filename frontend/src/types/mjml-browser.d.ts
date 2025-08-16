/**
 * Copyright 2025 JC-Lab
 * SPDX-License-Identifier: AGPL-3.0-or-later
 */

declare module 'mjml-browser' {
  const mjml2html: (mjml: string, opts?: any) => { html: string; errors?: any[] };
  export default mjml2html;
}
