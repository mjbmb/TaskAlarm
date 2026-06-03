Feature: Aufgaben-Übernahme vom Vortag
  Als Benutzer möchte ich gefragt werden ob unerledigte Aufgaben
  vom Vortag übernommen werden sollen,
  damit ich nichts vergesse.

  Background:
    Given ich habe keine Aufgaben
    And es gibt keine gestrigen Aufgaben

  Scenario: Keine gestrigen Aufgaben — keine Nachfrage
    Then gibt es keine übertragbaren Aufgaben vom Vortag

  Scenario: Erledigte Aufgaben vom Vortag werden nicht angeboten
    Given gestern gab es eine erledigte Aufgabe "Bericht fertig" mit Dauer "1h"
    Then gibt es keine übertragbaren Aufgaben vom Vortag

  Scenario: Ausstehende Aufgabe vom Vortag wird angeboten
    Given gestern gab es eine ausstehende Aufgabe "Offenes Ticket" mit Dauer "45m"
    Then gibt es 1 übertragbare Aufgabe vom Vortag
    And die übertragbare Aufgabe heißt "Offenes Ticket"

  Scenario: Aktive Aufgabe vom Vortag wird angeboten
    Given gestern gab es eine aktive Aufgabe "Halbfertiger Bericht" mit Dauer "2h"
    Then gibt es 1 übertragbare Aufgabe vom Vortag

  Scenario: Übernahme setzt Status auf ausstehend zurück
    Given gestern gab es eine aktive Aufgabe "Halbfertige Arbeit" mit Dauer "1h"
    When ich übernehme alle gestrigen Aufgaben
    Then enthält die Liste 1 Aufgabe
    And die erste Aufgabe ist ausstehend
    And die erste Aufgabe hat keine Startzeit

  Scenario: Mehrere gestrige Aufgaben werden korrekt angeboten
    Given gestern gab es eine ausstehende Aufgabe "Aufgabe 1" mit Dauer "30m"
    And gestern gab es eine erledigte Aufgabe "Aufgabe 2" mit Dauer "1h"
    And gestern gab es eine ausstehende Aufgabe "Aufgabe 3" mit Dauer "45m"
    Then gibt es 2 übertragbare Aufgaben vom Vortag

  Scenario: Übernahme fügt Aufgaben ans Ende der heutigen Liste
    Given ich füge eine Aufgabe "Heutige Aufgabe" mit Dauer "1h" hinzu
    And gestern gab es eine ausstehende Aufgabe "Gestrige Aufgabe" mit Dauer "30m"
    When ich übernehme alle gestrigen Aufgaben
    Then enthält die Liste 2 Aufgaben
    And die erste Aufgabe heißt "Heutige Aufgabe"
    And die zweite Aufgabe heißt "Gestrige Aufgabe"
