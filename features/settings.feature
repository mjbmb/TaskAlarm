Feature: Einstellungen
  Als Benutzer möchte ich Pufferzeit und Tagesende konfigurieren können,
  damit die Berechnung meinen Gewohnheiten entspricht.

  Background:
    Given ich habe keine Aufgaben

  Scenario: Standard-Pufferzeit ist 10 Minuten
    Then beträgt die Pufferzeit 10 Minuten

  Scenario: Standard-Tagesende ist 20:00 Uhr
    Then ist das Tagesende um 20:00 Uhr

  Scenario: Pufferzeit ändern
    When ich setze die Pufferzeit auf 20 Minuten
    Then beträgt die Pufferzeit 20 Minuten

  Scenario: Tagesende ändern
    When ich setze das Tagesende auf 18:30 Uhr
    Then ist das Tagesende um 18:30 Uhr

  Scenario: Pufferzeit wirkt sich auf Berechnung aus
    Given die aktuelle Zeit ist 10:00 Uhr
    And ich setze die Pufferzeit auf 30 Minuten
    And ich füge eine Aufgabe "Aufgabe A" mit Dauer "1h" hinzu
    And ich füge eine Aufgabe "Aufgabe B" mit Dauer "1h" hinzu
    Then liegt das voraussichtliche Ende um 12:30 Uhr

  Scenario: Früheres Tagesende macht Zeitplan enger
    Given die aktuelle Zeit ist 16:00 Uhr
    And ich setze das Tagesende auf 17:00 Uhr
    And ich füge eine Aufgabe "Aufgabe A" mit Dauer "2h" hinzu
    Then ist der Zeitplan überschritten
