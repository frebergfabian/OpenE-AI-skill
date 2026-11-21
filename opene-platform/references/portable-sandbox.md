# Portabel sandbox för Claude Code på Windows

Förhandsversion: `v0.5.0-dev`. Hämta från [GitHub Release](https://github.com/frebergfabian/OpenE-AI-skill/releases/tag/v0.5.0-dev).

Packa upp releasefilen `opene-platform.skill` (zip) till projektets `.claude/skills/` så att `.claude/skills/opene-platform/SKILL.md` finns. Paketet innehåller instruktioner och hjälpskript; exe-filen hämtas separat. Claude Code behöver lokal PowerShell och en webbläsare för UI-arbete.

Från skillmappen:

```powershell
powershell.exe -NoProfile -ExecutionPolicy Bypass -File sandbox/get-sandbox.ps1 -Version v0.5.0-dev -NoBrowser
$exe = Join-Path $env:LOCALAPPDATA 'OpenE-Sandbox/bin/OpenE-Sandbox-v0.5.0-dev-win-x64.exe'
& $exe status --json
& $exe info --secrets
& $exe stop
& $exe start --no-browser
```

`info --secrets` visar lokala slumpade lösenord för bland annat `sandbox-admin`, `handlaggare`, `sokande` och `sokande2`. Behåll dessa lokalt. Data lagras under `%LOCALAPPDATA%/OpenE-Sandbox`; `OPENE_SANDBOX_HOME` kan ange en egen katalog. Använd `--name testnamn` konsekvent för en separat instans. Starta inte olika exe-versioner mot samma instans.

Sex tjänster konfigureras automatiskt: Funktionslabbet, fastighet, barn, betalning, företag och flerpart. Simulator- och e-postadresser ges av `info`. Sparade ärenden granskas med respektive sökande och handläggare.

## Begränsningar

Detta är en testrelease. Samtliga ursprungliga acceptanskriterier är ännu inte uppfyllda. Separata utvecklingsprov har sparat bokning, kö, PUD, barnrelation, signering/flerpartssignering, dokument, HTTP, SMTP, SMS och simulerad betalning. Artvise CreateCase, Easit ImportItems och delar av SSBT har provats med genererade SOAP-klienter. Fullständiga leverantörs-API:er, alla verksamhetssystem och alla tillägg är inte implementerade eller verifierade. Läs [integrationsstatus](../sandbox/INTEGRATION-STATUS.md) och [portabel status](../sandbox/PORTABLE-STATUS.md).

Använd syntetiska data. Java-proxy och sanerade kända adresser utgör inte en fullständig nätverksspärr för råa socketar. Ren Windows-användare, samtliga tidskrav och återhämtning mitt i en betalningscallback är ännu inte fullständigt verifierade. Slutbyggets testresultat bifogas releasen.
