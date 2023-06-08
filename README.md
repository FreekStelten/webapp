# README
Dit is een eenvoudige webtoepassing die is geschreven in Go. Het biedt een gebruikersvriendelijke interface voor het opzoeken van voertuiggegevens op basis van kentekenplaten. De toepassing maakt gebruik van een Microsoft SQL Server-database voor het opslaan van de voertuiginformatie.

# Vereisten
- Go 1.15 of hoger
- Microsoft SQL Server-database
- Configuratiebestand (config.yaml) met de juiste databaseverbindinggegevens

# Installatie
1. Zorg ervoor dat Go is geïnstalleerd op uw systeem.
2. Maak een nieuwe directory en plaats het broncodebestand (main.go) in die directory.
3. Maak een configuratiebestand genaamd config.yaml in dezelfde directory met de volgende structuur:
-----------------------
- server: [servernaam]
- user_id: [gebruikersnaam]
- password: [wachtwoord]
- port: [poortnummer]
- database: [databasenaam]
-----------------------
Vervang [servernaam], [gebruikersnaam], [wachtwoord], [poortnummer] en [databasenaam] door de juiste databaseverbindinggegevens.

4. Open een terminal en navigeer naar de directory waar het programma is geplaatst.
5. oer het volgende commando uit om het programma te compileren en uit te voeren:

go run main.go

6. De webtoepassing wordt gestart en is beschikbaar op http://localhost:80.

# Gebruik
1. Open uw webbrowser en ga naar http://localhost:80.
2. U wordt gevraagd om in te loggen met een wachtwoord.
3. Voer het juiste wachtwoord in het daarvoor bestemde veld in.
4. Klik op de knop "Inloggen".
5. Als het wachtwoord correct is, wordt u doorgestuurd naar het zoekformulier.
6. Voer de kentekenplaat van het voertuig in het daarvoor bestemde veld in.
7. Klik op de knop "Zoeken".
8. De voertuiggegevens worden weergegeven op de pagina.

# Aanpassingen
Als u de gebruikersinterface wilt aanpassen of extra functionaliteit wilt toevoegen, kunt u het HTML-bestand index.html bewerken dat bij het programma wordt geleverd. U kunt ook de code in het bestand main.go aanpassen om aan uw specifieke vereisten te voldoen.

# Opmerking
Deze toepassing is bedoeld als een eenvoudige demonstratie en kan verder worden uitgebreid en aangepast voor meer geavanceerde functionaliteit en gebruikerservaring. Zorg ervoor dat u een veilig inlogsysteem implementeert als u deze toepassing in een productieomgeving wilt gebruiken.

# Foutmeldingen
Mogelijke foutmeldingen bij het gebruik van de toepassing:

"Fout bij het verbinden met de database"

Deze fout treedt op als de toepassing geen verbinding kan maken met de Microsoft SQL Server-database. Controleer of de database correct is geconfigureerd en bereikbaar is vanaf de toepassingshost.

"Ongeldig wachtwoord."
Deze fout treedt op als het ingevoerde wachtwoord niet overeenkomt met het verwachte wachtwoord voor inloggen. Controleer of u het juiste wachtwoord heeft ingevoerd.

"Kentekenplaat niet gevonden."
Deze fout treedt op als er geen voertuiggegevens beschikbaar zijn voor de ingevoerde kentekenplaat. Controleer of de ingevoerde kentekenplaat correct is en probeer het opnieuw.

Onbekende fout.
Deze fout treedt op als er een onbekende fout optreedt tijdens het opzoeken van de voertuiggegevens. Probeer het opnieuw en neem contact op met de beheerder als het probleem aanhoudt.