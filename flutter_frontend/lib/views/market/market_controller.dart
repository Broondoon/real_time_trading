import 'package:flutter/material.dart';

class MarketStateProvider extends ChangeNotifier {
  String _currStockId = '-1';
  String get getCurrStockId => _currStockId;

  String _currStockName = 'Google';
  String get getCurrStockName => _currStockName;

  String _currStockPrice = '999.99';
  String get getCurrStockPrice => _currStockPrice;

  void setStockShown(String id, String name, String price) {
    if (_currStockId != id) {
      _currStockId = id;
      _currStockName = name;
      _currStockPrice = price;
      notifyListeners();
    }
    else {
      print('Stock ID already shown.');
    }
  }
}