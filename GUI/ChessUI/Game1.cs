using Microsoft.Xna.Framework;
using Microsoft.Xna.Framework.Graphics;
using Microsoft.Xna.Framework.Input;
using System.Net.Sockets;
using System.Text.Json;
using System.Collections.Generic;
using System.Text;
using System;
using System.Diagnostics;
namespace ChessUI;

public class Game1 : Game
{
    private GraphicsDeviceManager _graphics;
    private SpriteBatch _spriteBatch;
    public static GameWindow gw;
  
    StringBuilder myTextBoxDisplayCharacters = new StringBuilder();
    Texture2D backgroundTexture;
    Texture2D ballTexture;
    Texture2D pawnTextureB;
    Texture2D pawnTextureW;
    Texture2D rookTextureB;
    Texture2D rookTextureW;
    Texture2D knightTextureB;
    Texture2D knightTextureW;
    Texture2D bishopTextureB;
    Texture2D bishopTextureW;
    Texture2D queenTextureB;
    Texture2D queenTextureW;
    Texture2D kingTextureB;
    Texture2D kingTextureW;
    Texture2D dotTexture;
    SpriteFont currentFont;
    Piece[][] boardArray;
   
    bool pToMove;
    string wCastleK;
    string wCastleQ;
    string bCastleK;
    string bCastleQ;
    string enPass;
    char[] squareLetter;
    int fullMoves;
    int halfMoves;
    FloatingPiece currentlyHeldPiece;
    Dictionary<int, Dictionary<int, int[][]>> legalMoves;
   // bool canStart;

    public static Process bockFish;

    public static int xOffset = 25;
    public static int yOffset = 10;
    public static int  squareSize = 86;
    public Game1()
    {
        _graphics = new GraphicsDeviceManager(this);
        Content.RootDirectory = "Content";
        IsMouseVisible = true;
        _graphics.PreferredBackBufferWidth = 720;
        _graphics.PreferredBackBufferHeight = 720;
        _graphics.ApplyChanges();
    }

    protected override void Initialize()
    {

        currentlyHeldPiece = null;

        pToMove = true;
        wCastleK = "K";
        wCastleQ = "Q";
        bCastleK = "k";
        bCastleQ = "q";
        enPass = " - ";

        squareLetter = new char[] 
        {
            'a',
            'b',
            'c',
            'd',
            'e',
            'f',
            'g',
            'h'
        };

        fullMoves = 0;


        bockFish = new Process();
        bockFish.StartInfo.FileName = "BockFish.exe";
        bockFish.StartInfo.UseShellExecute = false;
        bockFish.StartInfo.RedirectStandardInput = true;
        bockFish.StartInfo.RedirectStandardOutput = true;
        bockFish.Start();
        AppDomain.CurrentDomain.ProcessExit += closeBockFish;
        // Communicate with the other process
      

        sendCommand("uci");
        Console.WriteLine("Received: " + readAnswer());
        Console.WriteLine("Received: " + readAnswer());

        

        base.Initialize();
    }

    static void closeBockFish(object sender, EventArgs e){
        Console.WriteLine("Closing bockfish");
        bockFish.StandardInput.WriteLine("quit");
        bockFish.StandardInput.Close();
        bockFish.Close();
    }

    static void sendCommand(string command){
        bockFish.StandardInput.WriteLine(command);
    }

    static string readAnswer(){
        return bockFish.StandardOutput.ReadLine();
    }
    protected override void LoadContent()
    {
        _spriteBatch = new SpriteBatch(GraphicsDevice);
        gw = Window;

        // TODO: use this.Content to load your game content here
        backgroundTexture = Content.Load<Texture2D>("chessboard");
        ballTexture = Content.Load<Texture2D>("ball");
        pawnTextureB = Content.Load<Texture2D>("black-pawn");
        pawnTextureW = Content.Load<Texture2D>("white-pawn");
        rookTextureB = Content.Load<Texture2D>("black-rook");
        rookTextureW = Content.Load<Texture2D>("white-rook");
        knightTextureB = Content.Load<Texture2D>("black-knight");
        knightTextureW = Content.Load<Texture2D>("white-knight");
        bishopTextureB = Content.Load<Texture2D>("black-bishop");
        bishopTextureW = Content.Load<Texture2D>("white-bishop");
        queenTextureB = Content.Load<Texture2D>("black-queen");
        queenTextureW = Content.Load<Texture2D>("white-queen");
        kingTextureB = Content.Load<Texture2D>("black-king");
        kingTextureW = Content.Load<Texture2D>("white-king");
        dotTexture = Content.Load<Texture2D>("dot");
        currentFont = Content.Load<SpriteFont>("font");

        boardArray = new Piece[][] {
          new Piece[]{
            new Piece("r", rookTextureB, "b"), 
            new Piece("n", knightTextureB, "b"), 
            new Piece("b", bishopTextureB, "b"),
            new Piece("q", queenTextureB, "b"),
            new Piece("k", kingTextureB, "b"),
            new Piece("b", bishopTextureB, "b"),
            new Piece("n", knightTextureB, "b"),
            new Piece("r", rookTextureB, "b")},
            new Piece[]{
               new Piece("p", pawnTextureB, "b"), 
               new Piece("p", pawnTextureB, "b"), 
               new Piece("p", pawnTextureB, "b"), 
               new Piece("p", pawnTextureB, "b"), 
               new Piece("p", pawnTextureB, "b"), 
               new Piece("p", pawnTextureB, "b"), 
               new Piece("p", pawnTextureB, "b"), 
               new Piece("p", pawnTextureB, "b")},
            new Piece[]{ null, null, null, null, null, null, null, null },
            new Piece[]{ null, null, null, null, null, null, null, null },
            new Piece[]{ null, null, null, null, null, null, null, null },
            new Piece[]{ null, null, null, null, null, null, null, null },
            new Piece[]{
                new Piece("P", pawnTextureW, "w"), 
                new Piece("P", pawnTextureW, "w"), 
                new Piece("P", pawnTextureW, "w"), 
                new Piece("P", pawnTextureW, "w"), 
                new Piece("P", pawnTextureW, "w"), 
                new Piece("P", pawnTextureW, "w"), 
                new Piece("P", pawnTextureW, "w"), 
                new Piece("P", pawnTextureW, "w")},
          new Piece[]{
            new Piece("R", rookTextureW, "w"),
            new Piece("N", knightTextureW, "w"),
            new Piece("B", bishopTextureW, "w"),
            new Piece("Q", queenTextureW, "w"),
            new Piece("K", kingTextureW, "w"),
            new Piece("B", bishopTextureW, "w"),
            new Piece("N", knightTextureW, "w"),
            new Piece("R", rookTextureW, "w")
            }};
    }

    protected override void Update(GameTime gameTime)
    {
        if (GamePad.GetState(PlayerIndex.One).Buttons.Back == ButtonState.Pressed || Keyboard.GetState().IsKeyDown(Keys.Escape))
            Exit();

        // TODO: Add your update logic here

        var mstate = Mouse.GetState();
        var kstate = Keyboard.GetState();
      
        var isClicked = mstate.LeftButton == ButtonState.Pressed;
        if (isClicked)
        {
            System.Console.WriteLine("X: " + mstate.X + "  Y: "+ mstate.Y);
            currentlyHeldPiece = findClosestPiece(mstate);
            try
            {
                currentlyHeldPiece.pos.X = mstate.X;
                currentlyHeldPiece.pos.Y = mstate.Y; 
            }
            catch (System.Exception)
            {
                 
            }
        } else if (mstate.LeftButton == ButtonState.Released && currentlyHeldPiece != null)
        {
            dropPiece(currentlyHeldPiece, mstate);
        }

        base.Update(gameTime);
    }

    protected override void Draw(GameTime gameTime)
    {
        GraphicsDevice.Clear(Color.CornflowerBlue);

        // TODO: Add your drawing code here
        _spriteBatch.Begin();
        _spriteBatch.Draw(backgroundTexture, new Vector2(0, 0), Color.White); //Board
        drawBoard();
        if (currentlyHeldPiece != null)
        {
            drawFloatingPiece(currentlyHeldPiece);
        }
        _spriteBatch.End();

        base.Draw(gameTime);
    }

    void drawBoard() 
    {
        for (int y = 0; y < boardArray.Length; y++)
        {
            for (int x = 0; x < boardArray[y].Length; x++)
            {
                Piece piece = boardArray[y][x];
                if (piece != null)
                {

                    _spriteBatch.Draw(piece.sprite, 
                    new Vector2(squareSize*x + xOffset + squareSize/2, squareSize*y + yOffset + squareSize/2 ),
                    null, Color.White, 0f, new Vector2(piece.sprite.Width/2,piece.sprite.Height/2), Vector2.One, SpriteEffects.None, 0f);
                
                }
            }
        }
    }

    FloatingPiece findClosestPiece(MouseState mstate) 
    {
        if (currentlyHeldPiece == null) 
        {
          
            float x = System.MathF.Floor( (mstate.X - xOffset) / squareSize ); 
            float y = System.MathF.Floor( (mstate.Y - yOffset) / squareSize ) ;
            try{
                Piece closestPiece = boardArray[(int)y][(int)x];
                Vector2 origin = new Vector2(x,y);

                if (closestPiece != null) 
                {
                    boardArray[(int)y][(int)x] = null;
                    return new FloatingPiece(closestPiece.type, closestPiece.sprite, closestPiece.c, origin, origin);
                }
            }
            catch (System.Exception) {return null;}
        }
        return currentlyHeldPiece;
    }

    void dropPiece(FloatingPiece heldPiece, MouseState mstate)
    {
        Vector2 origin = heldPiece.origin;
        float x = System.MathF.Floor( (mstate.X - xOffset) / squareSize ) ;
        float y = System.MathF.Floor( (mstate.Y - yOffset) / squareSize ) ;
        bool validMove = true;
        try{
            for (int i = 0; i < legalMoves[(int)origin.Y][(int)origin.X].Length; i++) 
            {
                if (legalMoves[(int)origin.Y][(int)origin.X][i][0] == x && legalMoves[(int)origin.Y][(int)origin.X][i][1] == y)
                {
                    validMove = true;
                }
            }
        }
        catch (System.Exception){}
        
        if (validMove)
        {
          updateSpecialMoves(heldPiece, x, y);
         
          if (heldPiece.c == "b") fullMoves++;
          if (heldPiece.type.ToUpper() != "P" && boardArray[(int)y][(int)x] == null) halfMoves++; else halfMoves = 0;
          boardArray[(int)y][(int)x] = new Piece(heldPiece.type, heldPiece.sprite, heldPiece.c);
          currentlyHeldPiece = null;
          pToMove = !pToMove;
          Console.WriteLine(convertPosToFen());
        } else 
        {
            currentlyHeldPiece = null;
            boardArray[(int)origin.Y][(int)origin.X] = new Piece(heldPiece.type, heldPiece.sprite, heldPiece.c);
        }
    }

    void drawFloatingPiece(FloatingPiece piece)
    {
        /*
          try 
          {
              int boardWidth = _graphics.PreferredBackBufferWidth - offsetX;
              int boardHeight = _graphics.PreferredBackBufferHeight - offsetY;
              int[][] lMove = legalMoves[(int)System.MathF.Round(((currentlyHeldPiece.origin.Y + placeOffsetY) * 8 + 280) / boardHeight - 1)][(int)System.MathF.Round(((currentlyHeldPiece.origin.X + placeOffsetX) * 8 + 160) / boardWidth - 1)];
             
              foreach (int[] item in lMove)
              {
                  _spriteBatch.Draw(
                      dotTexture,
                      new Vector2(((boardWidth / 8) * (item[0] + 1)) - placeOffsetX, ((boardHeight / 8) * (item[1] + 1)) - placeOffsetY),
                      null,
                      Color.White,
                      0f,
                      new Vector2(dotTexture.Width / 2, dotTexture.Height / 2),
                      Vector2.One,
                      SpriteEffects.None,
                      0f
                      );
              }
          }
          catch (System.Exception)
          {
              
          }
        */
        _spriteBatch.Draw(
            piece.sprite,
            piece.pos,
            null,
            Color.White,
            0f,
            new Vector2(piece.sprite.Width / 2, piece.sprite.Height / 2),
            Vector2.One,
            SpriteEffects.None,
            0f
            );
    }

    void updateSpecialMoves(FloatingPiece heldPiece, float targetX, float targetY){
         switch (heldPiece.type)
          {
            case "K":
              wCastleK = wCastleQ = "";
              break;
            case "k":
              bCastleK = bCastleQ = "";
              break;

            case "R":
              if (heldPiece.origin.Y == 0)
              {
                  if (heldPiece.origin.X == 7)
                  {
                      wCastleK = "";
                  } else if (heldPiece.origin.X == 0) 
                  {
                      wCastleQ = "";
                  }
              }
              break;
            case "r":
              if (heldPiece.origin.Y == 7)
              {
                  if (heldPiece.origin.X == 7)
                  {
                      bCastleK = "";
                  } else if (heldPiece.origin.X == 0) 
                  {
                      bCastleQ = "";
                  }
              }
              break;

            case "P":
              if (targetY - heldPiece.origin.Y == -2) 
              {
                  enPass = " " + squareLetter[(int)targetX] + "3 ";
              }
              break;
            case "p":
              if (targetY - heldPiece.origin.Y == 2) 
              {
                  enPass = " " + squareLetter[(int)targetX] + "6 ";
              }
              break;
            default:
              break;
          }
    }

    string convertPosToFen()
    {
        string fen = "";
        int numOfEmpty;
        for (int y = 0; y < boardArray.Length; y++)
        {
            numOfEmpty = 0;
            for (int x = 0; x < boardArray[y].Length; x++)
            {
                Piece currentPiece = boardArray[y][x];
                if (currentPiece == null) 
                {
                    numOfEmpty++;
                } else
                {
                    if (numOfEmpty != 0) 
                    {
                        fen += numOfEmpty.ToString();
                        numOfEmpty = 0;
                    }
                    fen += currentPiece.type;
                }
            }
            if (numOfEmpty != 0) fen += numOfEmpty.ToString();
            fen += "/";
        }

        fen = fen.Remove(fen.Length - 1);
        fen += " ";
        if (pToMove) fen += "w "; else fen += "b ";

        fen += wCastleK;
        fen += wCastleQ;
        fen += bCastleK;
        fen += bCastleQ;
        if (wCastleK == "" && wCastleQ == "" && bCastleK == "" && bCastleQ == "") fen += "-";
        
        fen += enPass;
        enPass = " - ";

        fen += halfMoves.ToString() + " ";
        fen += fullMoves.ToString();

        return fen;
    }
}

class Piece
{
    public string type;
    public Texture2D sprite;
    public string c;
    public Piece(string _type, Texture2D texture, string color) 
    {
        type = _type;
        sprite = texture;
        c = color;
    }
}

class FloatingPiece : Piece
{
    public Vector2 pos;
    public Vector2 origin;
    public FloatingPiece(string type, Texture2D texture, string color, Vector2 position, Vector2 or) : base(type, texture, color) {
        pos = position;
        origin = or;
    }
}
