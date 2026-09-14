# Terms of Service

**Last Updated**: September 12, 2026

Welcome to DAWG (Digs Any Web-app Glitch), a developer tool for capturing, packaging, and reproducing web application states. By using the DAWG CLI, Desktop App, and related services (collectively, the "Software" or "Service"), you agree to these Terms of Service. You also represent that you are at least 18 years old or of legal age to form a binding contract in your jurisdiction.

## 1. Description of Service
DAWG is a developer utility that records selected browser-session data, sanitizes supported capture streams, and packages the result as a versioned OCI artifact. Artifacts may be kept in the local catalog, exported as portable `.dawg` archives, or explicitly shared through an OCI registry or another service.

## 2. User Responsibilities and Data Governance
- **Data Capture**: You acknowledge that DAWG can capture rrweb browser events, click and fill actions, frontend request metadata, and artifact metadata. Optional capture inputs may also include environment information, database data, structured logs, or HTTP cassettes.
- **Sensitive Information**: You are solely responsible for ensuring that you have the right to capture and share this data. You must not capture production data containing Personally Identifiable Information (PII), Protected Health Information (PHI), or highly sensitive credentials without appropriate authorization and safeguards.
- **Local Storage and Explicit Sharing**: DAWG stores captures and artifacts locally by default. The development team does not automatically receive them. If you export, push, upload, or otherwise share an artifact, the selected registry or third-party service may process that data under its own terms and privacy practices.
- **Sanitization and Review**: DAWG applies heuristic redaction and an OPA policy gate to supported files, but these controls do not guarantee removal of every sensitive value. You are responsible for reviewing artifacts and applying suitable organizational controls before distribution.

## 3. Acceptable Use
You agree not to:
- Use DAWG to capture data from applications you do not own, manage, or have explicit authorization to debug.
- Distribute DAWG artifacts containing malicious code, malware, or illicit content.
- Circumvent DAWG's replay, archive-validation, or other safety controls.

## 4. Intellectual Property
DAWG and its original content, features, and functionality are owned by its developers and are protected by international copyright, trademark, and other intellectual property or proprietary rights laws.

## 5. Open Source Licenses
The DAWG repository currently includes a root **GNU General Public License Version 3 (GPLv3)** license notice. DAWG also uses third-party dependencies distributed under their own licenses, including Apache License 2.0 and MIT terms.
- DAWG's original project code and each third-party component remain subject to their applicable open source license terms.
- Nothing in these Terms of Service restricts rights granted by, or supersedes obligations imposed by, an applicable open source license.

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
