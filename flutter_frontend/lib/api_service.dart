import 'dart:convert';

import 'package:dio/dio.dart';
import 'package:flutter_frontend/views/auth/auth_controller.dart';

class APIService {
  APIService(
    this._authController,
  );

  final AuthController _authController;
  final bool DEBUG_MODE = true;

  Future<Response> mockResponse(Map responseData) {
    Response<dynamic> response = Response<dynamic>(
      requestOptions: RequestOptions(path: ''),
      data: responseData,
      statusCode: 200,
    );

    return Future<Response<dynamic>>.value(
      response,
    );
  }

  //////////////////////////////////////////////
  
  Future<Response> getStockPrices() {
    if (!DEBUG_MODE) {
      final response = _authController.authdio.get(
        '/transaction/getStockPrices',
      );
      return response;
    }
    else {
      final responseData = {
        "success": true,
        "data": [
          {"stock_id":1,"stock_name":"Apple","current_price":100},
          {"stock_id":1, "stock_name":"Google","current_price":200}
        ]
      };

      return mockResponse(responseData);
    }
  }
  
  Future<Response> getWalletBalance() {
    if (!DEBUG_MODE) {
      final response = _authController.authdio.get(
        '/transaction/getWalletBalance',
      );
      return response;
    }
    else {
      final responseData = {
        "success": true,
        "data": [
          {"balance": 100},
        ]
      };

      return mockResponse(responseData);
    }
  }

  Future<Response> getPortfolio() {
    if (!DEBUG_MODE) {
      final response = _authController.authdio.get(
        '/transaction/getStockPortfolio',
      );
      return response;
    }
    else {
      final responseData = {
        "success": true,
        "data": [
          {"stock_id": 1, "stock_name": "Apple", "quantity_owned": 100},
          {"stock_id": 2, "stock_name": "Google", "quantity_owned": 150},
          {"stock_id": 3, "stock_name": "Apple", "quantity_owned": 100},
          {"stock_id": 4, "stock_name": "Google", "quantity_owned": 150},
          {"stock_id": 5, "stock_name": "Apple", "quantity_owned": 100},
          {"stock_id": 6, "stock_name": "Google", "quantity_owned": 150},
          {"stock_id": 7, "stock_name": "Apple", "quantity_owned": 100},
          {"stock_id": 8, "stock_name": "Google", "quantity_owned": 150},
          {"stock_id": 9, "stock_name": "Apple", "quantity_owned": 100},
          {"stock_id": 10, "stock_name": "Google", "quantity_owned": 150},
          {"stock_id": 11, "stock_name": "Apple", "quantity_owned": 100},
          {"stock_id": 12, "stock_name": "Google", "quantity_owned": 150},
          {"stock_id": 13, "stock_name": "Apple", "quantity_owned": 100},
          {"stock_id": 14, "stock_name": "Google", "quantity_owned": 150},
        ]
      };

      return mockResponse(responseData);
    }
  }

  Future<Response> getWalletTransactions() {
    if (!DEBUG_MODE) {
      final response = _authController.authdio.get(
        '/transaction/getWalletTransactions',
      );
      return response;
    }
    else {
      final responseData = {
        "success": true,
        "data": [
          {"wallet_tx_id":"628ba23df2210df6c3764823","stock_tx_id":"62738363a50350b1fbb243a6","is_debit":true,"amount":100,"time_stamp":"2025-01-12T15:03:25.019+00:00"}, 
          {"wallet_tx_id":"628ba36cf2210df6c3764824","stock_tx_id":"62738363a50350b1fbb243a6","is_debit":false,"amount":200,"time_stamp":"2025-0112T14:13:25.019+00:00"}
        ]
      };

      return mockResponse(responseData);
    }
  }

  Future<Response> getStockTransactions() {
    if (!DEBUG_MODE) {
      final response = _authController.authdio.get(
        '/transaction/StockTransactions',
      );
      return response;
    }
    else {
      final responseData = {
        "success": true,
        "data": [
          {
            "stock_tx_id":"62738363a50350b1fbb243a6","stock_id":1,
            "wallet_tx_id":"628ba23df2210df6c3764823","order_status":"COMPLETED",
            "is_buy":true,"order_type":"LIMIT","stock_price":50,"quantity":2,
            "parent_tx_id":null,"time_stamp":"2025-0112T15:03:25.019+00:00"
          }, 
          
          {
            "stock_tx_id":"62738363a50350b1fbb243a6","stock_id":1,
            "wallet_tx_id":"628ba36cf2210df6c3764824","order_status":"COMPLETED",
            "is_buy":false,"order_type":"MARKET","parent_tx_id":null, "stock_price":100,"quantity":2,
            "time_stamp":"2025-0112T14:13:25.019+00:00"
          }
        ]
      };

      return mockResponse(responseData);
    }
  }

  ////////////////////////////////////////////////////////////////////////////////////////////

  Future<Response> addMoneyToWallet(int amount) {
    if (!DEBUG_MODE) {
      Map reqData = {
        'amount': amount,
      };

      final response = _authController.authdio.post(
        '/transaction/addMoneyToWallet',
        data: jsonEncode(reqData),
      );
      return response;
    }
    else {
      final responseData = {
        "success": true,
        "data": null,
      };

      return mockResponse(responseData);
    }
  }

  Future<Response> placeStockOrder(
    int stockId, 
    bool isBuy,
    String orderType,
    int quantity,
    int price,
  ) {
    if (!DEBUG_MODE) {
      Map reqData = {
        "stock_id":stockId,
        "is_buy":isBuy,
        "order_type":orderType,
        "quantity":quantity,
        "price":price, 
      };

      final response = _authController.authdio.post(
        '/transaction/addMoneyToWallet',
        data: jsonEncode(reqData),
      );
      return response;
    }
    else {
      final responseData = {
        "success": true,
        "data": null,
      };

      return mockResponse(responseData);
    }
  }

  Future<Response> cancelStockTransaction(String stockTicketId) {
    if (!DEBUG_MODE) {
      Map reqData = {
        "stock_tx_id":stockTicketId,
      };

      final response = _authController.authdio.post(
        '/engine/cancelStockTransaction',
        data: jsonEncode(reqData),
      );
      return response;
    }
    else {
      final responseData = {
        "success": true,
        "data": null,
      };

      return mockResponse(responseData);
    }
  }

  ////////////////////////////////////////////////////////////////////////////////////////////

  Future<Response> createStock(String stockName) {
    if (!DEBUG_MODE) {
      Map reqData = {
        "stock_name":stockName,
      };

      final response = _authController.authdio.post(
        '/setup/createStock',
        data: jsonEncode(reqData),
      );
      return response;
    }
    else {
      final responseData = {
        "success": true,
        "data": null,
      };

      return mockResponse(responseData);
    }
  }

  Future<Response> addStockToUser(String stockId, int quantity) {
    if (!DEBUG_MODE) {
      Map reqData = {
        "stock_id":stockId,
        "quantity":quantity,
      };

      final response = _authController.authdio.post(
        '/setup/addStockToUser',
        data: jsonEncode(reqData),
      );
      return response;
    }
    else {
      final responseData = {
        "success": true,
        "data": null,
      };

      return mockResponse(responseData);
    }
  }

  ////////////////////////////////////////////////////////////////////////////////////////////

  // The context is needed to build a SnackBar popup! But it's a pain to pass around the context, and
  //    implementing a dedicated service is a low priority rn.
  // And until I do that, or provide some other unique behaviour to this class, using this is pointless.
  // So... TODO: do that.
  // Future<Response?> post(context, String path, {dynamic data}) async {
  // Future<Response?> post(String path, {dynamic data}) async {

  //   try {
  //     final response = await _authController.authdio.post(
  //       path,
  //       data: data,
  //     );
  //     return response;
  //   }
  //   on DioException catch (e) {

  //     print(">>> ALERT! $e");

  //     // if (!context.mounted) {
  //     //   print(">>> ALERT! Could not display exception to use: $e");
  //     //   return null;
  //     // }

  //     // ScaffoldMessenger.of(context).showSnackBar(
  //     //   SnackBar(
  //     //     content: Text('Request Failure: $e'),
  //     //   ),
  //     // );

  //     return null;
  //   }
  //   catch (e) {
  //     print('>>> Something really bad happened. Generic exception: $e');
  //     rethrow;
  //   }
  // }

  // Future<Response?> get(context, String path) async {
  // Future<Response?> get(String path) async {
  //   try {
  //     final response = await _authController.authdio.get(path);
  //     return response;
  //   }
  //   on Exception catch (e) {

  //     print(">>> ALERT! $e");

  //     // if (!context.mounted) {
  //     //   print(">>> ALERT! Could not display exception to use: $e");
  //     //   return null;
  //     // }

  //     // ScaffoldMessenger.of(context).showSnackBar(
  //     //   SnackBar(
  //     //     content: Text('Request Failure: $e'),
  //     //   ),
  //     // );

  //     return null;
  //   }
  // }
}

// TODO: in the future, see if standardizing a Response class that I control is useful.
// class BaseResponse<T> {
//   final int statusCode;
//   final String message;
//   final T? data;
//   // final bool success;

//   bool get isSuccessful => statusCode >= 200 && statusCode < 300;

//   BaseResponse({
//     required this.statusCode,
//     required this.message,
//     this.data,
//     // isSuccessful,
//   });

    // Pretend as if a factory .fromJSON is here.
// }