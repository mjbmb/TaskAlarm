Feature: Erinnerungen und Alarme
  Als Benutzer möchte ich akustisch erinnert werden wenn ich Aufgaben vergesse
  oder meinen Zeitplan überschreite.

  Background:
    Given ich habe keine Aufgaben
    And das Tagesende ist um 20:00 Uhr
    And die Pufferzeit beträgt 0 Minuten
    And es wurde noch keine Erinnerung ausgelöst

  Scenario: Erinnerung wenn ausstehende Aufgabe vorhanden
    Given die aktuelle Zeit ist 10:00 Uhr
    And ich füge eine Aufgabe "Offene Aufgabe" mit Dauer "1h" hinzu
    When der Erinnerungs-Check läuft
    Then wird eine Erinnerung ausgelöst
    And kein Alarm wird ausgelöst

  Scenario: Kein Alarm wenn Zeitplan eingehalten wird
    Given die aktuelle Zeit ist 10:00 Uhr
    And ich füge eine Aufgabe "Kleine Aufgabe" mit Dauer "1h" hinzu
    When der Erinnerungs-Check läuft
    Then wird keine Alarm ausgelöst

  Scenario: Alarm wenn Zeitplan überschritten
    Given die aktuelle Zeit ist 19:00 Uhr
    And ich füge eine Aufgabe "Große Aufgabe" mit Dauer "3h" hinzu
    When der Erinnerungs-Check läuft
    Then wird ein Alarm ausgelöst
    And keine Erinnerung wird ausgelöst

  Scenario: Keine Erinnerung innerhalb von 10 Minuten nach letzter
    Given die aktuelle Zeit ist 10:00 Uhr
    And ich füge eine Aufgabe "Aufgabe" mit Dauer "1h" hinzu
    And der Erinnerungs-Check lief vor 5 Minuten
    When der Erinnerungs-Check läuft
    Then wird keine Erinnerung ausgelöst

  Scenario: Erinnerung nach genau 10 Minuten erneut
    Given die aktuelle Zeit ist 10:00 Uhr
    And ich füge eine Aufgabe "Aufgabe" mit Dauer "1h" hinzu
    And der Erinnerungs-Check lief vor 10 Minuten
    When der Erinnerungs-Check läuft
    Then wird eine Erinnerung ausgelöst

  Scenario: Kein Alarm innerhalb von 10 Minuten nach letztem
    Given die aktuelle Zeit ist 19:00 Uhr
    And ich füge eine Aufgabe "Große Aufgabe" mit Dauer "3h" hinzu
    And der Alarm lief vor 5 Minuten
    When der Erinnerungs-Check läuft
    Then wird keine Alarm ausgelöst

  Scenario: Kein Erinnerung außerhalb der Arbeitszeit — zu früh
    Given die aktuelle Zeit ist 06:00 Uhr
    And ich füge eine Aufgabe "Aufgabe" mit Dauer "1h" hinzu
    When der Erinnerungs-Check läuft
    Then wird keine Erinnerung ausgelöst
    And wird keine Alarm ausgelöst

  Scenario: Keine Erinnerung außerhalb der Arbeitszeit — zu spät
    Given die aktuelle Zeit ist 21:30 Uhr
    And ich füge eine Aufgabe "Aufgabe" mit Dauer "1h" hinzu
    When der Erinnerungs-Check läuft
    Then wird keine Erinnerung ausgelöst

  Scenario: Keine Erinnerung wenn alle Aufgaben erledigt
    Given die aktuelle Zeit ist 10:00 Uhr
    And ich füge eine Aufgabe "Erledigte Aufgabe" mit Dauer "1h" hinzu
    And ich starte die Aufgabe "Erledigte Aufgabe"
    And ich schließe die Aufgabe "Erledigte Aufgabe" ab
    When der Erinnerungs-Check läuft
    Then wird keine Erinnerung ausgelöst
    And wird keine Alarm ausgelöst

  Scenario: Alarm hat Priorität über Erinnerung
    Given die aktuelle Zeit ist 19:00 Uhr
    And ich füge eine Aufgabe "Große Aufgabe" mit Dauer "3h" hinzu
    When der Erinnerungs-Check läuft
    Then wird ein Alarm ausgelöst
    And keine Erinnerung wird ausgelöst
