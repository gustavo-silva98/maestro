from selenium.webdriver.firefox.service import Service
from selenium import webdriver
from selenium.webdriver.common.by import By
from selenium.webdriver.support.ui import WebDriverWait
from selenium.webdriver.support import expected_conditions as EC

import os
import time
import pandas as pd
import pyscreenshot

def newDriver(driver_exec:str) -> webdriver.Firefox:
    service = Service()
    driver = webdriver.Firefox(service=service)
    driver.maximize_window()
    return driver

def readCSV(csv_name:str) -> pd.DataFrame:
    return pd.read_csv(f"{csv_name}.csv",sep=";")

def aceitar_cookie(driver:webdriver.Chrome) -> bool:
    WebDriverWait(driver,10).until(EC.presence_of_element_located((By.TAG_NAME,"iframe")))
    iframe = driver.find_elements(By.TAG_NAME,"iframe")
    for frame in iframe:
        if frame.get_attribute("title") == "Iframe title":
            driver.switch_to.frame(frame)
            buttons = driver.find_elements(By.TAG_NAME,"button")
            for x in buttons:
                if "Aceitar e continuar" in x.text:
                    x.click()
                    driver.switch_to.default_content()
                    return True
    return False

def find_player_transfermarkt(driver:webdriver.Chrome,player_name:str) -> bool:
    url = f"https://www.transfermarkt.com.br/schnellsuche/ergebnis/schnellsuche?query={player_name}"
    driver.get(url=url)
    time.sleep(6)
    aceitar_cookie(driver=driver)
    class_search = driver.find_elements(By.CLASS_NAME,"hauptlink")
    for classes in class_search:
        if classes.text == player_name:
            classes.click()
            break
    if driver.save_screenshot(f"Prints/{player_name}.png"):
        return True
    else:
        return False
    



def printa_usuario(img_name:str) -> bool:
    
    #WebDriverWait(driver, 10).until(EC.element_to_be_clickable((By.CSS_SELECTOR,'input#ctl00_Cph_pnlGeral_ctl00_UcCadUsuarioWebUnificado_janela_jnlDadosGerais_txtNome_CAMPO')))
    image = pyscreenshot.grab()
    image.save(f"{img_name}.jpeg") # type: ignore
    return True

def main():
    if not (os.path.isdir('Prints')):
        os.mkdir('Prints')
    input = readCSV("input")
    driver = newDriver("chromedriver.exe")
    for index, row in enumerate(input.itertuples()):
        find_player_transfermarkt(driver=driver,player_name=str(row.Atleta))
    


main()