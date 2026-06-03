Feature: Zeitplan-Berechnung
  Als Benutzer möchte ich sehen ob ich meine Aufgaben bis zum Tagesende schaffe,
  damit ich meinen Tag realistisch planen kann.

  Background:
    Given ich habe keine Aufgaben
    And das Tagesende ist um 20:00 Uhr
    And die Pufferzeit beträgt 0 Minuten

  Scenario: Alle Aufgaben passen in den Tag
    Given die aktuelle Zeit ist 10:00 Uhr
    And ich füge eine Aufgabe "Aufgabe A" mit Dauer "2h" hinzu
    And ich füge eine Aufgabe "Aufgabe B" mit Dauer "3h" hinzu
    Then liegt das voraussichtliche Ende um 15:00 Uhr
    And ist der Zeitplan im grünen Bereich

  Scenario: Aufgaben passen nicht mehr in den Tag
    Given die aktuelle Zeit ist 18:00 Uhr
    And ich füge eine Aufgabe "Aufgabe A" mit Dauer "3h" hinzu
    Then liegt das voraussichtliche Ende um 21:00 Uhr
    And ist der Zeitplan überschritten

  Scenario: Pufferzeit wird zwischen Aufgaben addiert
    Given die aktuelle Zeit ist 10:00 Uhr
    And die Pufferzeit beträgt 15 Minuten
    And ich füge eine Aufgabe "Aufgabe A" mit Dauer "1h" hinzu
    And ich füge eine Aufgabe "Aufgabe B" mit Dauer "1h" hinzu
    Then liegt das voraussichtliche Ende um 12:15 Uhr

  Scenario: Erledigte Aufgaben werden nicht mitgerechnet
    Given die aktuelle Zeit ist 10:00 Uhr
    And ich füge eine Aufgabe "Fertige Aufgabe" mit Dauer "5h" hinzu
    And ich starte die Aufgabe "Fertige Aufgabe"
    And ich schließe die Aufgabe "Fertige Aufgabe" ab
    And ich füge eine Aufgabe "Offene Aufgabe" mit Dauer "1h" hinzu
    Then liegt das voraussichtliche Ende um 11:00 Uhr
    And ist der Zeitplan im grünen Bereich

  Scenario: Aktive Aufgabe reduziert verbleibende Zeit
    Given die aktuelle Zeit ist 10:00 Uhr
    And ich füge eine Aufgabe "Laufende Aufgabe" mit Dauer "2h" hinzu
    And ich starte die Aufgabe "Laufende Aufgabe" um 10:00 Uhr
    When 30 Minuten vergehen
    Then liegt das voraussichtliche Ende um 12:00 Uhr

  Scenario: Konfigurierbares Tagesende
    Given die aktuelle Zeit ist 17:00 Uhr
    And das Tagesende ist um 19:00 Uhr
    And ich füge eine Aufgabe "Aufgabe A" mit Dauer "3h" hinzu
    Then ist der Zeitplan überschritten

  Scenario: Exakt am Limit
    Given die aktuelle Zeit ist 18:00 Uhr
    And ich füge eine Aufgabe "Letzte Aufgabe" mit Dauer "2h" hinzu
    Then liegt das voraussichtliche Ende um 20:00 Uhr
    And ist der Zeitplan im grünen Bereich
