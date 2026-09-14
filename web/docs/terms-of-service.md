# Terms of Service

**Last Updated**: September 12, 2026

Welcome to DAWG (Digs Any Web-app Glitch), a developer tool for capturing, packaging, and reproducing web application states. By using the DAWG CLI, Desktop App, and related services (collectively, the "Software" or "Service"), you agree to these Terms of Service. You also represent that you are at least 18 years old or of legal age to form a binding contract in your jurisdiction.

## 1. Description of Service
DAWG is a developer utility that packages complete bug reproduction states—including environment variables, application state, network requests, and input sequences—into a single portable, versioned OCI artifact. The software operates primarily on the user's local machine and creates artifacts intended for local or shared debugging.

## 2. User Responsibilities and Data Governance
- **Data Capture**: You acknowledge that DAWG captures deep application state, including database fixtures (`application/vnd.dawg.db.fixture+tar.zstd`), network requests/responses (`application/vnd.dawg.cassette+tar.zstd`), UI interaction traces (e.g., `rrweb` logs), and environment configurations.
- **Sensitive Information**: You are solely responsible for ensuring that you have the right to capture and share this data. You must not capture production data containing Personally Identifiable Information (PII), Protected Health Information (PHI), or highly sensitive credentials without strict anonymization.
- **No Data Collection by DAWG Developers**: All data and artifacts are recorded and stored 100% locally on your machine. The DAWG development team does not collect, receive, or have any access to the data you capture.
- **Sanitization & Sharing Requirements**: Because artifacts are stored locally, you are solely responsible when sharing (exporting) the files with third parties. DAWG provides a deny-by-default Open Policy Agent (OPA) sanitization engine, but since we have no access to your local artifacts, we are not liable for any data leaks if you choose to distribute them.

## 3. Acceptable Use
You agree not to:
- Use DAWG to capture data from applications you do not own, manage, or have explicit authorization to debug.
- Distribute DAWG artifacts containing malicious code, malware, or illicit content.
- Circumvent the built-in sandbox-only replay constraints (e.g., bypassing `dawg run` safety checks).

## 4. Intellectual Property
DAWG and its original content, features, and functionality are owned by its developers and are protected by international copyright, trademark, and other intellectual property or proprietary rights laws.

## 5. Open Source Licenses
Portions of the DAWG software are made available under open source licenses, including the **GNU General Public License, Version 3, 29 June 2007 (GPLv3)** and the **Apache License, Version 2.0 (Apache 2.0)**.
- The rights and obligations concerning the use, modification, and distribution of DAWG's source code are governed strictly by those respective open source licenses. 
- Nothing in these Terms of Service restricts your rights under, or grants you rights that supersede, the terms of any applicable open source license.

## 6. Disclaimer of Warranty
THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED. WE DO NOT WARRANT THAT THE SANITIZATION ENGINE WILL CATCH ALL SENSITIVE DATA. THE ENTIRE RISK ARISING OUT OF THE USE OF THE SOFTWARE REMAINS WITH YOU.

## 7. Limitation of Liability
IN NO EVENT SHALL DAWG, ITS CREATORS, OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, OR CONSEQUENTIAL DAMAGES (INCLUDING DATA BREACHES OR EXPOSURE OF SENSITIVE INFORMATION) ARISING OUT OF THE USE OR INABILITY TO USE THE SOFTWARE.

## 8. Changes to Terms
We reserve the right to modify these Terms at any time. We will notify users of significant changes by updating the documentation repository.

## 9. Indemnification
You agree to indemnify, defend, and hold harmless DAWG, its creators, and contributors from and against any and all claims, liabilities, damages, losses, and expenses (including reasonable legal fees) arising out of or in any way connected with your misuse of the Software, your violation of these Terms, or your violation of any third-party rights (including data theft or privacy breaches).

## 10. Governing Law & Jurisdiction
These Terms shall be governed by and construed in accordance with the laws of the Republic of Indonesia. Any disputes arising from or relating to the use of the Software shall be subject to the exclusive jurisdiction of the courts of the Republic of Indonesia.
