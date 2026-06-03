Feature: Aufgabenverwaltung
  Als Benutzer möchte ich Aufgaben verwalten können,
  um meinen Arbeitstag zu planen.

  Background:
    Given ich habe keine Aufgaben

  Scenario: Aufgabe hinzufügen
    When ich füge eine Aufgabe "E-Mails beantworten" mit Dauer "30m" hinzu
    Then enthält die Liste 1 Aufgabe
    And die erste Aufgabe heißt "E-Mails beantworten"
    And die erste Aufgabe hat eine Dauer von 30 Minuten
    And die erste Aufgabe ist ausstehend

  Scenario: Mehrere Aufgaben hinzufügen
    When ich füge eine Aufgabe "Meeting" mit Dauer "1h" hinzu
    And ich füge eine Aufgabe "Bericht schreiben" mit Dauer "2h" hinzu
    Then enthält die Liste 2 Aufgaben
    And die erste Aufgabe heißt "Meeting"
    And die zweite Aufgabe heißt "Bericht schreiben"

  Scenario: Aufgabe entfernen
    Given ich füge eine Aufgabe "Unwichtige Aufgabe" mit Dauer "15m" hinzu
    When ich entferne die Aufgabe "Unwichtige Aufgabe"
    Then enthält die Liste 0 Aufgaben

  Scenario: Aufgabe nach oben verschieben
    Given ich füge eine Aufgabe "Erste" mit Dauer "15m" hinzu
    And ich füge eine Aufgabe "Zweite" mit Dauer "15m" hinzu
    When ich verschiebe "Zweite" nach oben
    Then ist die erste Aufgabe "Zweite"
    And ist die zweite Aufgabe "Erste"

  Scenario: Aufgabe nach unten verschieben
    Given ich füge eine Aufgabe "Erste" mit Dauer "15m" hinzu
    And ich füge eine Aufgabe "Zweite" mit Dauer "15m" hinzu
    When ich verschiebe "Erste" nach unten
    Then ist die erste Aufgabe "Zweite"
    And ist die zweite Aufgabe "Erste"

  Scenario: Aufgabe starten
    Given ich füge eine Aufgabe "Tagesplanung" mit Dauer "20m" hinzu
    When ich starte die Aufgabe "Tagesplanung"
    Then ist die Aufgabe "Tagesplanung" aktiv

  Scenario: Aufgabe abschließen
    Given ich füge eine Aufgabe "Tagesplanung" mit Dauer "20m" hinzu
    And ich starte die Aufgabe "Tagesplanung"
    When ich schließe die Aufgabe "Tagesplanung" ab
    Then ist die Aufgabe "Tagesplanung" erledigt

  Scenario: Dauereingabe als Minuten-Zahl
    When ich füge eine Aufgabe "Kurze Aufgabe" mit Dauer "45" hinzu
    Then die erste Aufgabe hat eine Dauer von 45 Minuten

  Scenario: Dauereingabe als Stunden und Minuten
    When ich füge eine Aufgabe "Lange Aufgabe" mit Dauer "1h30m" hinzu
    Then die erste Aufgabe hat eine Dauer von 90 Minuten
